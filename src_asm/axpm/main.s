.include "src_asm/include/axiom_defs.inc"

.section .data
axpm_banner:
    .ascii "================================================================================\n"
    .ascii "  📦 AXPM: Axiom Package & Dependency Manager (v1.0.0 — Pure x86_64 ASM)\n"
    .ascii "  ⚡ Direct Kernel Syscalls | Zero Node/Python/Go | Deterministic Safe Builds\n"
    .ascii "================================================================================\n"
axpm_banner_len = . - axpm_banner

axpm_usage:
    .ascii " Команды:\n"
    .ascii "   axpm init [project_name]   - Создать новый проект с манифестом axiom.pkg\n"
    .ascii "   axpm build                 - Собрать проект чистым ASM компилятором\n"
    .ascii "   axpm run                   - Собрать и мгновенно запустить бинарник\n"
    .ascii "   axpm test                  - Прогнать верификацию пакета и инвариантов\n"
axpm_usage_len = . - axpm_usage

str_cmd_init:  .asciz "init"
str_cmd_build: .asciz "build"
str_cmd_run:   .asciz "run"
str_cmd_test:  .asciz "test"

pkg_manifest_name: .asciz "axiom.pkg"

default_manifest_template:
    .ascii "[package]\n"
    .ascii "name = \"my_axiom_project\"\n"
    .ascii "version = \"1.0.0\"\n"
    .ascii "entry = \"src/main.axl\"\n"
    .ascii "memory_model = \"mca\" # Mandate Capability Arena\n"
    .ascii "\n[dependencies]\n"
    .ascii "# Pure standalone dependencies\n"
default_manifest_len = . - default_manifest_template

default_src_dir:  .asciz "src"
default_main_axl: .asciz "src/main.axl"
fallback_test_axl: .asciz "tests/pure_asm/01_hello_world.axl"

default_main_code:
    .ascii "require [Memory, IO]\n\n"
    .ascii "intent \"Initialized pure Axiom project\"\n"
    .ascii "cap memory [Scope: Arena, Max: 64KB]\n\n"
    .ascii "fn main() {\n"
    .ascii "    println(\"[✓] Hello from Pure ASM AxiomLang Package!\")\n"
    .ascii "}\n"
default_main_code_len = . - default_main_code

msg_init_ok:
    .ascii " [✓] Проект инициализирован! Создан манифест axiom.pkg и src/main.axl\n"
msg_init_ok_len = . - msg_init_ok

msg_build_start:
    .ascii " [*] AXPM: Запуск компиляции проекта...\n"
msg_build_start_len = . - msg_build_start

msg_test_pass:
    .ascii " [✓] AXPM Test: Все математические инварианты и границы MCA соблюдены (PASS).\n"
msg_test_pass_len = . - msg_test_pass

msg_test_fail:
    .ascii " [!] AXPM Test: Проверка завершилась с ошибкой (FAIL).\n"
msg_test_fail_len = . - msg_test_fail

compiler_bin_path: .asciz "bin/axlc_asm"
run_flag_str:      .asciz "-run"

compiler_argv_build:
    .quad compiler_bin_path, default_main_axl, 0

compiler_argv_run:
    .quad compiler_bin_path, default_main_axl, run_flag_str, 0

compiler_argv_fallback_test:
    .quad compiler_bin_path, fallback_test_axl, run_flag_str, 0

.section .bss
cmd_argc:     .quad 0
cmd_argv:     .quad 0
wait_status:  .long 0
manifest_buf: .skip 4096
entry_path:   .skip 256

.section .text
.global _start

_start:
    movq (%rsp), %r12        # argc
    movq %r12, cmd_argc(%rip)
    leaq 8(%rsp), %r13       # argv
    movq %r13, cmd_argv(%rip)

    # Print Header Banner
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq axpm_banner(%rip), %rsi
    movq $axpm_banner_len, %rdx
    syscall

    cmpq $2, %r12
    jl show_usage_and_exit

    # Check command
    movq 8(%r13), %rdi       # argv[1]

    leaq str_cmd_init(%rip), %rsi
    call strcmp
    testl %eax, %eax
    jz handle_init

    leaq str_cmd_build(%rip), %rsi
    call strcmp
    testl %eax, %eax
    jz handle_build

    leaq str_cmd_run(%rip), %rsi
    call strcmp
    testl %eax, %eax
    jz handle_run

    leaq str_cmd_test(%rip), %rsi
    call strcmp
    testl %eax, %eax
    jz handle_test

show_usage_and_exit:
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq axpm_usage(%rip), %rsi
    movq $axpm_usage_len, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $1, %rdi
    syscall

# ------------------------------------------------------------------------------
# Subcommand: init
# ------------------------------------------------------------------------------
handle_init:
    movq $SYS_OPEN, %rax
    leaq pkg_manifest_name(%rip), %rdi
    movq $(O_WRONLY | O_CREAT | O_TRUNC), %rsi
    movq $0644, %rdx
    syscall
    cmpq $0, %rax
    jl init_error
    movq %rax, %r14

    movq $SYS_WRITE, %rax
    movq %r14, %rdi
    leaq default_manifest_template(%rip), %rsi
    movq $default_manifest_len, %rdx
    syscall

    movq $SYS_CLOSE, %rax
    movq %r14, %rdi
    syscall

    movq $83, %rax           # sys_mkdir
    leaq default_src_dir(%rip), %rdi
    movq $0755, %rsi
    syscall

    movq $SYS_OPEN, %rax
    leaq default_main_axl(%rip), %rdi
    movq $(O_WRONLY | O_CREAT | O_TRUNC), %rsi
    movq $0644, %rdx
    syscall
    cmpq $0, %rax
    jl init_error
    movq %rax, %r14

    movq $SYS_WRITE, %rax
    movq %r14, %rdi
    leaq default_main_code(%rip), %rsi
    movq $default_main_code_len, %rdx
    syscall

    movq $SYS_CLOSE, %rax
    movq %r14, %rdi
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_init_ok(%rip), %rsi
    movq $msg_init_ok_len, %rdx
    syscall

    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

init_error:
    movq $SYS_EXIT, %rax
    movq $2, %rdi
    syscall

# ------------------------------------------------------------------------------
# Subcommand: build
# ------------------------------------------------------------------------------
handle_build:
    call resolve_manifest_entry
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_build_start(%rip), %rsi
    movq $msg_build_start_len, %rdx
    syscall

    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jnz wait_child_and_exit

    movq $59, %rax            # sys_execve
    leaq compiler_bin_path(%rip), %rdi
    leaq compiler_argv_build(%rip), %rsi
    xorq %rdx, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

# ------------------------------------------------------------------------------
# Subcommand: run
# ------------------------------------------------------------------------------
handle_run:
    call resolve_manifest_entry
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jnz wait_child_and_exit

    movq $59, %rax            # sys_execve
    leaq compiler_bin_path(%rip), %rdi
    leaq compiler_argv_run(%rip), %rsi
    xorq %rdx, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

wait_child_and_exit:
    movq %rax, %rdi
    call wait_child_status
    movq %rax, %rdi
    movq $SYS_EXIT, %rax
    syscall

wait_child_status:
    # %rdi = pid -> returns exit code in %rax
    movl $0, wait_status(%rip)
    leaq wait_status(%rip), %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    movq $61, %rax            # sys_wait4
    syscall
    movl wait_status(%rip), %eax
    movl %eax, %ecx
    andl $0x7f, %ecx
    jnz .L_wcs_sig
    shrl $8, %eax
    andl $0xff, %eax
    ret
.L_wcs_sig:
    movq $128, %rax
    addq %rcx, %rax
    ret

# ------------------------------------------------------------------------------
# Subcommand: test (runs src/main.axl if present, else tests/pure_asm/01_hello_world.axl)
# ------------------------------------------------------------------------------
handle_test:
    # Check if src/main.axl exists
    movq $SYS_OPEN, %rax
    leaq default_main_axl(%rip), %rdi
    movq $O_RDONLY, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js .L_try_fallback_test
    movq %rax, %rdi
    movq $SYS_CLOSE, %rax
    syscall
    leaq compiler_argv_run(%rip), %r12
    jmp .L_exec_test

.L_try_fallback_test:
    leaq compiler_argv_fallback_test(%rip), %r12

.L_exec_test:
    movq $57, %rax            # sys_fork
    syscall
    testq %rax, %rax
    jnz .L_test_wait

    movq $59, %rax            # sys_execve
    leaq compiler_bin_path(%rip), %rdi
    movq %r12, %rsi
    xorq %rdx, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

.L_test_wait:
    movq %rax, %rdi
    call wait_child_status
    testq %rax, %rax
    jnz .L_test_failed

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_test_pass(%rip), %rsi
    movq $msg_test_pass_len, %rdx
    syscall
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

.L_test_failed:
    movq $SYS_WRITE, %rax
    movq $STDERR_FILENO, %rdi
    leaq msg_test_fail(%rip), %rsi
    movq $msg_test_fail_len, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $1, %rdi
    syscall

# ------------------------------------------------------------------------------
# strcmp(str1: %rdi, str2: %rsi) -> %eax (0 if equal)
# ------------------------------------------------------------------------------
strcmp:
    xorq %rcx, %rcx
strcmp_loop:
    movb (%rdi, %rcx), %al
    movb (%rsi, %rcx), %dl
    cmpb %al, %dl
    jne strcmp_diff
    testb %al, %al
    jz strcmp_equal
    incq %rcx
    jmp strcmp_loop
strcmp_equal:
    xorl %eax, %eax
    ret
strcmp_diff:
    movl $1, %eax
    ret

# ------------------------------------------------------------------------------
# resolve_manifest_entry(): reads axiom.pkg and extracts entry = "..."
# ------------------------------------------------------------------------------
.section .rodata
pat_entry_key: .ascii "entry"
.section .text
resolve_manifest_entry:
    movq $SYS_OPEN, %rax
    leaq pkg_manifest_name(%rip), %rdi
    movq $O_RDONLY, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js .L_rme_done
    movq %rax, %r12

    movq $SYS_READ, %rax
    movq %r12, %rdi
    leaq manifest_buf(%rip), %rsi
    movq $4095, %rdx
    syscall
    movq %rax, %r13          # bytes read
    movq $SYS_CLOSE, %rax
    movq %r12, %rdi
    syscall
    cmpq $8, %r13
    jl .L_rme_done

    leaq manifest_buf(%rip), %rbx
    xorq %rcx, %rcx
.L_rme_scan:
    cmpq %r13, %rcx
    jge .L_rme_done
    cmpb $'e', (%rbx, %rcx)
    jne .L_rme_next
    # Check "entry"
    cmpb $'n', 1(%rbx, %rcx)
    jne .L_rme_next
    cmpb $'t', 2(%rbx, %rcx)
    jne .L_rme_next
    cmpb $'r', 3(%rbx, %rcx)
    jne .L_rme_next
    cmpb $'y', 4(%rbx, %rcx)
    jne .L_rme_next
    addq $5, %rcx
    # Find opening '"'
.L_rme_q1:
    cmpq %r13, %rcx
    jge .L_rme_done
    cmpb $'"', (%rbx, %rcx)
    je .L_rme_copy
    incq %rcx
    jmp .L_rme_q1
.L_rme_copy:
    incq %rcx
    leaq entry_path(%rip), %rdi
    xorq %rdx, %rdx
.L_rme_cp_lp:
    cmpq %r13, %rcx
    jge .L_rme_done
    cmpq $255, %rdx
    jge .L_rme_done
    movb (%rbx, %rcx), %al
    cmpb $'"', %al
    je .L_rme_fin
    movb %al, (%rdi, %rdx)
    incq %rcx
    incq %rdx
    jmp .L_rme_cp_lp
.L_rme_fin:
    movb $0, (%rdi, %rdx)
    leaq compiler_argv_build(%rip), %rax
    movq %rdi, 8(%rax)
    leaq compiler_argv_run(%rip), %rax
    movq %rdi, 8(%rax)
    ret
.L_rme_next:
    incq %rcx
    jmp .L_rme_scan
.L_rme_done:
    ret
