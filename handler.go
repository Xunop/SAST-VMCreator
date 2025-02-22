package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

var configChan = make(chan map[string]string, 1)

func processCommands(ctx context.Context, q *CommandQueue) {
	for {
		cmd := q.Dequeue()
		if cmd != nil {
			go handleCommand(ctx, *cmd)
		}
		time.Sleep(time.Second)
	}
}

func handleCommand(ctx context.Context, cmd Command) {
	switch cmd.Type {
	case "/create_vm":
		handleCreateVM(ctx, cmd)
		// Add more command handlers as needed
	case "/help":
		handleHelp(ctx, cmd)
	case "/release":
	    handleRelease(ctx, cmd)
	}
}

func handleRelease(ctx context.Context, cmd Command) {
	messageID, ok := activeTopics.Load("current_message_id")
	if !ok {
		fmt.Println("Failed to get message_id from context")
		return
	}
	threadID, ok := activeTopics.Load("current_thread_id")
	if !ok {
		fmt.Println("Failed to get thread_id from context")
		return
	}
	curThreadID := cmd.Event.Message.ThreadID
	if threadID.(string) != curThreadID {
		_, err := sendReply(ctx, messageID.(string), fmt.Sprintf("<at user_id=\"%s\">%s</at> Please release the lock by replying to the at bot /release command within this thread!", cmd.Event.Sender.UserID, cmd.Event.Sender.UserID), true)
		if err != nil {
			fmt.Println("Failed to send reply:", err)
		}
		return
	}
	terraformMutex.Unlock()
	sendReply(ctx, cmd.Event.Message.MessageID, "Lock released", false)
}

func handleHelp(ctx context.Context, cmd Command) {
	_, err := sendReply(ctx, cmd.Event.Message.MessageID, HelpMsg, false)
	if err != nil {
		fmt.Println("Failed to send help message:", err)
	}
}

// Global mutex to ensure only one Terraform deployment runs at a time
var terraformMutex sync.Mutex

func handleCreateVM(ctx context.Context, cmd Command) {
	if !terraformMutex.TryLock() {
		_, err := sendReply(ctx, cmd.Event.Message.MessageID, "another Terraform deployment is running", false)
		if err != nil {
			fmt.Println("Failed to send reply:", err)
		}
		return
	}
	msgRsp, err := sendReply(ctx, cmd.Event.Message.MessageID, ExampleConfig, true)
	if err != nil {
		fmt.Println("Failed to send reply:", err)
		terraformMutex.Unlock() // Release lock if reply fails
		return
	}

	// Store topic information
	activeTopics.Store(msgRsp.ThreadID, &TopicInfo{
		UserID:   cmd.Event.Sender.UserID,
		RootID:   cmd.Event.Message.MessageID,
		ParentID: msgRsp.MessageID, // Parent ID is the message where the example config is sent
		ThreadID: msgRsp.ThreadID,
	})
	// Store current topic thread id in context
	ctx = context.WithValue(ctx, "thread_id", msgRsp.ThreadID)
	ctx = context.WithValue(ctx, "message_id", msgRsp.MessageID)
	activeTopics.Store("current_message_id", msgRsp.MessageID)
	activeTopics.Store("current_thread_id", msgRsp.ThreadID)

	select {
	case <-time.After(5 * time.Minute):
		// Remove the topic if no reply is received within 5 minutes
		activeTopics.Delete(msgRsp.ThreadID)
		terraformMutex.Unlock()
		_, err := sendReply(ctx, msgRsp.MessageID, "Configuration timeout. Please try again.", true)
		if err != nil {
			fmt.Println("Failed to send timeout message:", err)
			terraformMutex.Unlock() // Release lock if reply fails
			return
		}
		return
	case userConf := <-configChan:
		err := applyTerraformConfig(ctx, userConf)
		// If an error occurs, send a failure message
		if err != nil {
			fmt.Println("Failed to apply Terraform configuration:", err)
			terraformMutex.Unlock() // Ensure the mutex is released in case of error
			sendReply(ctx, msgRsp.MessageID, "Failed to create VM. Please try again.", true)
			return
		}
		// If the configuration is successfully applied, remove the topic
		activeTopics.Delete(msgRsp.ThreadID)
		terraformMutex.Unlock()
		_, err = sendReply(ctx, msgRsp.MessageID, fmt.Sprintf("<at user_id=\"%s\">%s</at> VM successfully created!", cmd.Event.Sender.UserID, cmd.Event.Sender.UserID), true)
		// _, err = sendReply(ctx, msgRsp.MessageID, "VM successfully created!", true)
		if err != nil {
			fmt.Println("Failed to send success message:", err)
			return
		}
	}

	return
}

// Listen for replies within a specific topic
func handleReply(ctx context.Context, cmd Command) error {
	event := cmd.Event
	// Ensure we are processing replies within an active topic
	_, exists := activeTopics.Load(event.Message.ThreadID)
	if !exists {
		return nil // Message is not part of an active `/create_vm` topic
	}

	message := event.Message
	if !message.ContainesBotMention() {
		return nil
	}

	// Apply Terraform Configuration
	config := parseConfig(message.Content.Text)
	// Send the configuration to the handleCreateVM function
	configChan <- config

	return nil
}

func sendReply(ctx context.Context, messageID, content string, replyInThread bool) (*MessageResponse, error) {
	client := lark.NewClient(AppID, AppSecret)

	// Properly format the content into JSON
	contentMap := map[string]string{
		"text": content,
	}
	contentBytes, err := json.Marshal(contentMap)
	if err != nil {
		return nil, err
	}

	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			Content(string(contentBytes)).
			MsgType("text").
			ReplyInThread(replyInThread).
			Uuid(generateUUID()). // Generate a new UUID for each request
			Build()).
		Build()

	resp, err := client.Im.Message.Reply(ctx, req)
	if err != nil {
		return nil, err
	}

	msgResp, err := mapToMessageResponse(resp)
	if err != nil {
		return nil, err
	}

	return msgResp, nil
}

func parseConfig(configStr string) map[string]string {
	config := make(map[string]string)
	lines := strings.Split(configStr, "\n")
	for _, line := range lines {
		// Ignore empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Use regex to correctly split by the first '='
		pair := regexp.MustCompile(`^([^=]+)=(.*)$`).FindStringSubmatch(line)
		if len(pair) == 3 {
			key := strings.TrimSpace(pair[1])
			value := strings.TrimSpace(pair[2])

			// If value is enclosed in quotes, remove the quotes
			if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				value = strings.Trim(value, `"`)
			}

			// If value is in format [IP](URL), extract the IP address
			ipPattern := regexp.MustCompile(`^\[(.*?)\]\(.*?\)$`)
			ipMatch := ipPattern.FindStringSubmatch(value)
			if len(ipMatch) == 2 {
				value = ipMatch[1] // Extract the IP address inside the brackets
			}

			config[key] = value
		}
	}
	return config
}

func applyTerraformConfig(ctx context.Context, config map[string]string) error {
	// fmt.Println("Applying Terraform configuration:", config)

	// Retrieve the thread ID from the context and create the working directory
	threadID, ok := ctx.Value("thread_id").(string)
	if !ok {
		return fmt.Errorf("thread_id not found in context")
	}
	dirPath := filepath.Join("generate", threadID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	defer clearUp(ctx)
	fmt.Println("Running Terraform in directory:", dirPath)

	// Define paths for required files and create symbolic links
	files := map[string]string{
		"terraform/main.tf":             "main.tf",
		"terraform/variable.tf":         "variable.tf",
		"terraform/.terraform":          ".terraform",
		"terraform/.terraform.lock.hcl": ".terraform.lock.hcl",
		"cloud-init/userdata.yaml":      "userdata.yaml",
	}
	for src, dest := range files {
		absSrc, err := filepath.Abs(src)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", src, err)
		}
		if err := createSymlink(absSrc, filepath.Join(dirPath, dest)); err != nil {
			return err
		}
	}

	// Write the terraform.tfvars file with configuration values
	if err := writeTfVarsFile(filepath.Join(dirPath, "terraform.tfvars"), config); err != nil {
		return err
	}

	// Run Terraform commands with a timeout context
	terraformCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := runTerraformCommand(terraformCtx, dirPath, "init"); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	if err := runTerraformCommand(terraformCtx, dirPath, "apply", "-auto-approve"); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	// Retrieve Terraform outputs
	ips, err := getTerraformOutputIPs(terraformCtx, dirPath)
	if err != nil {
		return fmt.Errorf("failed to retrieve Terraform outputs: %w", err)
	}
	messageID, ok := ctx.Value("message_id").(string)
	if !ok {
		return fmt.Errorf("message_id not found in context")
	}
	// Send a success message with all ip addresses, use neline to separate them
	_, err = sendReply(ctx, messageID, "VM successfully created with IP addresses:\n"+strings.Join(ips, "\n"), true)
	if err != nil {
		return fmt.Errorf("failed to send success message: %w", err)
	}

	return nil
}

// clearUp cleans up the working directory after Terraform deployment(wheather success or failure)
func clearUp(ctx context.Context) error {
	threarID, ok := ctx.Value("thread_id").(string)
	if !ok {
		return fmt.Errorf("thread_id not found in context")
	}

	dirPath := filepath.Join("generate", threarID)
	return os.RemoveAll(dirPath)
}

// createSymlink safely creates a symbolic link
func createSymlink(src, dest string) error {
	if err := os.Symlink(src, dest); err != nil {
		return fmt.Errorf("failed to create symlink from %s to %s: %w", src, dest, err)
	}
	return nil
}

// writeTfVarsFile writes key-value pairs to a terraform.tfvars file
func writeTfVarsFile(path string, config map[string]string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create terraform.tfvars: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for key, value := range config {
		if _, err := writer.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, value)); err != nil {
			return fmt.Errorf("failed to write to terraform.tfvars: %w", err)
		}
	}
	return writer.Flush()
}

// runTerraformCommand executes a Terraform command in the specified directory
func runTerraformCommand(ctx context.Context, dirPath, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, "terraform", append([]string{command}, args...)...)
	cmd.Dir = dirPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getTerraformOutputIPs retrieves the 'ip' output from Terraform
func getTerraformOutputIPs(ctx context.Context, dirPath string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "terraform", "output", "-json")
	cmd.Dir = dirPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute terraform output: %w", err)
	}

	var outputs map[string]struct {
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(output, &outputs); err != nil {
		return nil, fmt.Errorf("failed to parse terraform output JSON: %w", err)
	}

	ipOutput, exists := outputs["ip"]
	if !exists {
		return nil, fmt.Errorf("output 'ip' not found in Terraform outputs")
	}

	// Assuming 'ip' is a list of strings
	ipList, ok := ipOutput.Value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for 'ip' output")
	}

	var ips []string
	for _, ip := range ipList {
		ipStr, ok := ip.(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for IP address: %v", ip)
		}
		ips = append(ips, ipStr)
	}

	return ips, nil
}

// Generate UUID.
func generateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

// copyFile copy a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return os.Chtimes(dst, sourceInfo.ModTime(), sourceInfo.ModTime())
}

// copyDir copy a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(entries))
	concurrency := 10
	sem := make(chan struct{}, concurrency)

	for _, entry := range entries {
		wg.Add(1)
		sem <- struct{}{}

		go func(entry os.DirEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())

			if entry.IsDir() {
				if err := copyDir(srcPath, dstPath); err != nil {
					errCh <- err
				}
			} else {
				if err := copyFile(srcPath, dstPath); err != nil {
					errCh <- err
				}
			}
		}(entry)
	}

	wg.Wait()
	close(errCh)

	for copyErr := range errCh {
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}
