
#!/usr/bin/env bash
# Executing this script:
# curl https://raw.githubusercontent.com/ColDog/kubedeploy/master/hack/install.sh | sudo bash


REPO="ColDog/kubedeploy"
CMD="kubedeploy"
INSTALL_DIR=${INSTALL_DIR:="/usr/local/bin"}

# initArch discovers the architecture for this system.
initArch() {
  ARCH=$(uname -m)
  case $ARCH in
    armv5*) ARCH="armv5";;
    armv6*) ARCH="armv6";;
    armv7*) ARCH="armv7";;
    aarch64) ARCH="arm64";;
    x86) ARCH="386";;
    x86_64) ARCH="amd64";;
    i686) ARCH="386";;
    i386) ARCH="386";;
  esac
}

# initOS discovers the operating system for this system.
initOS() {
  OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

  case "$OS" in
    # Minimalist GNU for Windows
    mingw*) OS='windows';;
  esac
}

# initVersion will grep the latest version.
initVersion() {
  VERSION=$( curl -sSL "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
  )
}

# verifySupported checks that the os/arch combination is supported for
# binary builds.
verifySupported() {
  local supported="darwin-386\ndarwin-amd64\nlinux-386\nlinux-amd64\nlinux-arm\nlinux-arm64\nwindows-386\nwindows-amd64"
  if ! echo "${supported}" | grep -q "${OS}-${ARCH}"; then
    echo "No prebuilt binary for ${OS}-${ARCH}."
    echo "To build from source, go to https://github.com/$REPO"
    exit 1
  fi

  if ! type "curl" > /dev/null; then
    echo "curl is required"
    exit 1
  fi
}

# fail_trap is executed if an error occurs.
failTrap() {
  result=$?
  if [ "$result" != "0" ]; then
    echo "Failed to install $CMD"
  fi
  exit $result
}

# downloadCmd will download the cmd from github.
downloadCmd() {
  echo "Fetching release $CMD-$VERSION-$OS-$ARCH..."
  curl -sSL \
    https://github.com/$REPO/releases/download/$VERSION/$CMD-$VERSION-$OS-$ARCH \
    -o $INSTALL_DIR/$CMD
  chmod +x $INSTALL_DIR/$CMD
  echo "Installed successfully into $INSTALL_DIR/$CMD"
}

trap "failTrap" EXIT

initArch
initOS
verifySupported
initVersion
downloadCmd
