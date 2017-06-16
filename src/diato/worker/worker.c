// Diato - Reverse Proxying for Hipsters
//
// Copyright 2016-2017 Dolf Schimmel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

void isolateFileSystem(void) {
  char chroot_dir[256] = "";
  size_t len;
  // This is awfully Linux-specific. Other (more Posix-like) OS's should
  // probably use fcntl with F_GETPATH
  if ((len = readlink("/proc/self/fd/3", chroot_dir, 255)) == -1) {
    fprintf(stderr, "Could not determine chroot directory: %s\n",
            strerror(errno));
  } else {
    chroot_dir[len] = '\0';
  }

  if (chroot(chroot_dir) != 0) {
    fprintf(stderr, "Could not chroot() worker: %s\n", strerror(errno));
    exit(1);
  }

  if (chdir("/") != 0) {
    fprintf(stderr, "Could not chdir('/') worker: %s\n", strerror(errno));
    exit(1);
  }

  return;
}

void dropUserPrivs(void) {
  if (setgid(65534) != 0) {
    fprintf(stderr, "Could not setgid() to group nobody: %s\n",
            strerror(errno));
    exit(1);
  }

  if (setuid(65534) != 0) {
    fprintf(stderr, "Could not setuid() to user nobody: %s\n", strerror(errno));
    exit(1);
  }

  return;
}

void secureEnvironment(void) {

  if (fdopen(3, "r") == 0) {
    // We're the master process
    return;
  }

  isolateFileSystem();
  dropUserPrivs();

  // TODO: Drop capabilities?
}
