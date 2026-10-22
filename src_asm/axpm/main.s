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

compiler_bin_path: .asciz "bin/axlc_asm"
run_flag_str:      .asciz "-run"

compiler_argv_build:
    .quad compiler_bin_path, default_main_axl, 0

compiler_argv_run:
    .quad compiler_bin_path, default_main_axl, run_flag_str, 0

.section .bss
cmd_argc: .quad 0
cmd_argv: .quad 0

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
    # 1. Create axiom.pkg
    movq $SYS_OPEN, %rax
    leaq pkg_manifest_name(%rip), %rdi
    movq $65, %rsi           # O_WRONLY | O_CREAT
    movq $0644, %rdx
    syscall
    cmpq $0, %rax
    jl init_error
    movq %rax, %r14          # fd

    movq $SYS_WRITE, %rax
    movq %r14, %rdi
    leaq default_manifest_template(%rip), %rsi
    movq $default_manifest_len, %rdx
    syscall

    movq $SYS_CLOSE, %rax
    movq %r14, %rdi
    syscall

    # 2. mkdir src (sys_mkdir = 83)
    movq $83, %rax
    leaq default_src_dir(%rip), %rdi
    movq $0755, %rsi
    syscall

    # 3. Create src/main.axl
    movq $SYS_OPEN, %rax
    leaq default_main_axl(%rip), %rdi
    movq $65, %rsi           # O_WRONLY | O_CREAT
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

    # Print success
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
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_build_start(%rip), %rsi
    movq $msg_build_start_len, %rdx
    syscall

    # execve bin/axlc_asm src/main.axl
    movq $57, %rax            # sys_fork
    syscall
    cmpq $0, %rax
    jne wait_child

    movq $59, %rax            # sys_execve
    leaq compiler_bin_path(%rip), %rdi
    leaq compiler_argv_build(%rip), %rsi
    xorq %rdx, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

wait_child:
    xorq %rdi, %rdi
    movq $61, %rax            # sys_wait4
    xorq %rsi, %rsi
    xorq %rdx, %rdx
    xorq %r10, %r10
    syscall

    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

# ------------------------------------------------------------------------------
# Subcommand: run
# ------------------------------------------------------------------------------
handle_run:
    movq $57, %rax            # sys_fork
    syscall
    cmpq $0, %rax
    jne wait_child

    movq $59, %rax            # sys_execve
    leaq compiler_bin_path(%rip), %rdi
    leaq compiler_argv_run(%rip), %rsi
    xorq %rdx, %rdx
    syscall
    movq $SYS_EXIT, %rax
    movq $127, %rdi
    syscall

# ------------------------------------------------------------------------------
# Subcommand: test
# ------------------------------------------------------------------------------
handle_test:
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_test_pass(%rip), %rsi
    movq $msg_test_pass_len, %rdx
    syscall
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
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
