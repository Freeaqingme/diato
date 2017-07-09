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

#include <sys/prctl.h>
//#include <linux/seccomp.h>
#include <seccomp.h>

#include <sys/capability.h>

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
    fprintf(stderr, "Could not chroot() worker into '%s': %s\n", chroot_dir, strerror(errno));
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

// TODO: Sort of a PoC that definitely needs more refinement
// http://web.archive.org/web/20160429070024/http://www.lorier.net/docs/dropping-privs
int dropCaps(void) {
  cap_t no_cap;

  no_cap = cap_init();

  if (cap_clear(no_cap) == -1) {
    cap_free(no_cap);
    return -1;
  }
  if (cap_set_proc(no_cap) == -1) {
    cap_free(no_cap);
  }

  cap_free(no_cap);
  return 0; /* Success */
}

// TODO: Is clearly not yet done
// https://gist.github.com/sbz/1090868
void sandbox(void) {
  scmp_filter_ctx ctx;
  ctx = seccomp_init(SCMP_ACT_ALLOW);

//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigreturn), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(read), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(write), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(sched_getaffinity), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mmap), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(munmap), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigprocmask), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(sigaltstack), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(gettid), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigaction), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mprotect), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(clone), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(futex), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(set_robust_list), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(readlinkat), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(pselect6), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(sched_yield), 0);
//  seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(accept), 0);

  seccomp_load(ctx);
}

void secureEnvironment(void) {

  if (fdopen(4, "r") == 0) {
    // We're the master process
    return;
  }

  //isolateFileSystem(); // TODO
  dropUserPrivs();
  dropCaps();
//  sandbox(); // TODO
}
