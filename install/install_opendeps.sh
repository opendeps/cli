#!/usr/bin/env bash

# Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# OpenDeps CLI installation
# See: https://github.com/opendeps/cli

set -e

BASE_URL="https://github.com/opendeps/cli/releases/download"
LATEST_RELEASE_API="https://api.github.com/repos/opendeps/cli/releases/latest"

function unsupported_arch() {
  echo "This OS/architecture is unsupported."
  exit 1
}

function is_macos() {
  case "$(uname -s)" in
  *darwin* ) true ;;
  *Darwin* ) true ;;
  * ) false;;
  esac
}

function is_linux() {
  case "$(uname -s)" in
  *Linux* ) true ;;
  *linux* ) true ;;
  * ) false;;
  esac
}

function find_arch() {
  if [[ is_macos ]]; then
    case "$(uname -p)" in
    *i386* ) OPENDEPS_ARCH="x86_64" ;;
    *x86_64* ) OPENDEPS_ARCH="x86_64" ;;
    * ) unsupported_arch;;
    esac
  else
    case "$(uname -p)" in
    *i686* ) OPENDEPS_ARCH="x86_64" ;;
    *x86_64* ) OPENDEPS_ARCH="x86_64" ;;
    *armv6* ) OPENDEPS_ARCH="arm" ;;
    *armv7* ) OPENDEPS_ARCH="arm" ;;
    *arm64* ) OPENDEPS_ARCH="arm64" ;;
    * ) unsupported_arch;;
    esac
  fi
}

function find_os() {
    if [[ is_macos ]]; then
      OPENDEPS_OS="macOS"
    elif [[ is_linux ]]; then
      OPENDEPS_OS="Linux"
    else
      unsupported_arch
    fi
}

function find_version() {
    if [[ -z "${OPENDEPS_VERSION}" ]]; then
      echo "Attempting to determine latest version..."
      if [[ ! $( command -v jq ) ]]; then
        echo "Error: jq must be installed on your system in order to determine latest version."
        echo "Either install jq or set OPENDEPS_VERSION."
        exit 1
      fi

      OPENDEPS_VERSION="$( curl --fail --silent "${LATEST_RELEASE_API}" | jq -c '.tag_name' --raw-output )"
    fi

    if [[ "${OPENDEPS_VERSION:0:1}" == "v" ]]; then
      OPENDEPS_VERSION="$( echo ${OPENDEPS_VERSION} | cut -c 2- )"
    fi
    echo "Using version: ${OPENDEPS_VERSION}"
}

find_os
find_arch
find_version
DOWNLOAD_URL="${BASE_URL}/v${OPENDEPS_VERSION}/opendeps_${OPENDEPS_VERSION}_${OPENDEPS_OS}_${OPENDEPS_ARCH}.tar.gz"

OPENDEPS_TEMP_DIR="$( mktemp -d /tmp/opendeps.XXXXXXX )"
cd "${OPENDEPS_TEMP_DIR}"

echo -e "\nDownloading from ${DOWNLOAD_URL}"
curl --fail -L -o opendeps.tar.gz "${DOWNLOAD_URL}"
tar xf opendeps.tar.gz

echo -e "\nInstalling to /usr/local/bin"
cp ./opendeps /usr/local/bin/opendeps

echo -e "\nDone"
