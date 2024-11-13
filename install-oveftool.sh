#!/bin/bash

set -e

# ----------------------------------------------------------------
# 1. Define Variables
# ----------------------------------------------------------------

PKGNAME="vmware-ovftool"
PKGVER="4.6.3.24031167"
PKGREL="1"
URL="https://developer.broadcom.com/tools/open-virtualization-format-ovf-tool/latest"
DEPENDS=("expat" "openssl" "xerces-c" "icu" "libxcrypt-compat")

# Since we can't download from the original URL, we'll use a local mirror
DOWNLOAD_URL="http://192.168.114.4:54331/images/VMware-ovftool-4.6.3-24031167-lin.x86_64.zip"
# Donwload icu60 from local mirror
ICU60URL="http://192.168.114.4:54331/images/icu60-60.3-1-x86_64.pkg.tar.zst"

# Construct the source ZIP filename based on pkgver
SOURCE_ZIP="VMware-ovftool-${PKGVER%.*}-${PKGVER##*.}-lin.x86_64.zip"

# Expected SHA256 checksum of the source ZIP file
SHA256SUM_EXPECTED="344c30826121fe6ad990c32123e2aaf9261d777309d3e2012dd521775fa43e8c"

# ----------------------------------------------------------------
# 2. Check Dependencies
# ----------------------------------------------------------------

echo "Checking dependencies..."
MISSING_DEPS=()

for dep in "${DEPENDS[@]}"; do
    pacman -Q "$dep" >/dev/null 2>&1 || MISSING_DEPS+=("$dep")
done

# Install ICU60 and MISSING_DEPS
echo "Downloading ICU60..."
wget -O "icu60-60.3-1-x86_64.pkg.tar.zst" "$ICU60URL"
pacman -U "icu60-60.3-1-x86_64.pkg.tar.zst" --noconfirm
rm -f "icu60-60.3-1-x86_64.pkg.tar.zst"
if [ ${#MISSING_DEPS[@]} -ne 0 ]; then
    echo "The following dependencies are missing: ${MISSING_DEPS[*]}"
    echo "Please install them using 'pacman -S' and retry."
    exit 1
else
    echo "All dependencies are satisfied."
fi

# for dep in "${DEPENDS[@]}"; do
#     if ! apk info -e "$dep" >/dev/null 2>&1; then
#         MISSING_DEPS+=("$dep")
#     fi
# done

# if [ ${#MISSING_DEPS[@]} -ne 0 ]; then
#     echo "The following dependencies are missing: ${MISSING_DEPS[*]}"
#     echo "Please install them using 'apk add' and retry."
#     exit 1
# else
#     echo "All dependencies are satisfied."
# fi

# ----------------------------------------------------------------
# 3. Download the Source File
# ----------------------------------------------------------------

echo "Downloading source file from $DOWNLOAD_URL..."
wget -O "$SOURCE_ZIP" "$DOWNLOAD_URL"

# ----------------------------------------------------------------
# 4. Verify Source File
# ----------------------------------------------------------------

echo "Verifying source file: $SOURCE_ZIP"
if [[ ! -f "$SOURCE_ZIP" ]]; then
    echo "Source file '$SOURCE_ZIP' not found. Please place it in the script's directory."
    exit 1
fi

echo "Calculating SHA256 checksum..."
SHA256SUM_ACTUAL=$(sha256sum "$SOURCE_ZIP" | awk '{print $1}')
if [[ "$SHA256SUM_ACTUAL" != "$SHA256SUM_EXPECTED" ]]; then
    echo "SHA256 checksum does not match!"
    echo "Expected: $SHA256SUM_EXPECTED"
    echo "Actual:   $SHA256SUM_ACTUAL"
    exit 1
fi
echo "Checksum verification passed."

# ----------------------------------------------------------------
# 5. Extract Source File
# ----------------------------------------------------------------

TEMP_DIR=$(mktemp -d)
echo "Extracting '$SOURCE_ZIP' to temporary directory '$TEMP_DIR'..."
unzip -q "$SOURCE_ZIP" -d "$TEMP_DIR"

OVFTOOL_DIR="$TEMP_DIR/ovftool"
if [[ ! -d "$OVFTOOL_DIR" ]]; then
    echo "After extraction, 'ovftool' directory not found."
    exit 1
fi

# ----------------------------------------------------------------
# 6. Install Files
# ----------------------------------------------------------------

echo "Installing files to system directories..."

# Install binaries and libraries
INSTALL_DIR="/usr/lib/${PKGNAME}"
mkdir -p "$INSTALL_DIR"

echo "Copying binaries and libraries to '$INSTALL_DIR'..."
cp -v "$OVFTOOL_DIR/libgoogleurl.so.59" \
      "$OVFTOOL_DIR/libssoclient.so" \
      "$OVFTOOL_DIR/libvmacore.so" \
      "$OVFTOOL_DIR/libvmomi.so" \
      "$OVFTOOL_DIR/libvim-types.so" \
      "$OVFTOOL_DIR/ovftool" \
      "$OVFTOOL_DIR/ovftool.bin" \
      "$INSTALL_DIR/"

chmod 755 "$INSTALL_DIR/libgoogleurl.so.59" \
       "$INSTALL_DIR/libssoclient.so" \
       "$INSTALL_DIR/libvmacore.so" \
       "$INSTALL_DIR/libvmomi.so" \
       "$INSTALL_DIR/libvim-types.so" \
       "$INSTALL_DIR/ovftool" \
       "$INSTALL_DIR/ovftool.bin"

# Install data files
DATA_SUBDIRS=("certs" "env" "env/en" "schemas/DMTF" "schemas/vmware")
for subdir in "${DATA_SUBDIRS[@]}"; do
    TARGET_DIR="$INSTALL_DIR/$subdir"
    mkdir -p "$TARGET_DIR"
    cp -vr "$OVFTOOL_DIR/$subdir/"* "$TARGET_DIR/"
    chmod 755 "$TARGET_DIR"
    chmod 644 "$TARGET_DIR/"*.*
done

# Create symbolic link for the main script
BIN_DIR="/usr/bin"
mkdir -p "$BIN_DIR"
ln -sf "/usr/lib/${PKGNAME}/ovftool" "$BIN_DIR/ovftool"
chmod 755 "$BIN_DIR/ovftool"

# Install license files
LICENSE_DIR="/usr/share/licenses/${PKGNAME}"
mkdir -p "$LICENSE_DIR"
cp -v "$OVFTOOL_DIR/vmware.eula" \
      "$OVFTOOL_DIR/vmware-eula.rtf" \
      "$OVFTOOL_DIR/open_source_licenses.txt" \
      "$LICENSE_DIR/"
chmod 755 "$LICENSE_DIR"
chmod 644 "$LICENSE_DIR/"*.txt "$LICENSE_DIR/"*.eula "$LICENSE_DIR/"*.rtf

# Install documentation files
DOC_DIR="/usr/share/doc/${PKGNAME}"
mkdir -p "$DOC_DIR"
cp -v "$OVFTOOL_DIR/README.txt" "$DOC_DIR/"
chmod 755 "$DOC_DIR"
chmod 644 "$DOC_DIR/README.txt"

# ----------------------------------------------------------------
# 7. Clean Up Temporary Files
# ----------------------------------------------------------------

echo "Cleaning up temporary files..."
rm -rf "$TEMP_DIR"

echo "Installation of '${PKGNAME}' version ${PKGVER}-${PKGREL} completed successfully!"
