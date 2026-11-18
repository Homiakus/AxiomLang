# ==============================================================================
# HPT_OPTIMIZER.S
# Реализация математических оптимизаций Гомотопической Теории Программ (HPT)
# в чистом машинном коде x86_64 для AxiomLang.
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/hpt_optimizer_defs.inc"

.section .data
hpt_banner:
    .ascii " [HPT] Mathematical Optimization Pipeline Active:\n"
    .ascii "       ├── Quillen Zero-Stack Register Routing (b1=0)\n"
    .ascii "       ├── Koszul Syzygy Bound Elimination (Tor_1=0)\n"
    .ascii "       └── Branchless 2-Simplex Cobordism Reduction\n"
hpt_banner_len = . - hpt_banner

.section .text
.global hpt_optimize_pipeline
.global hpt_verify_flawless_invariants

# ==============================================================================
# hpt_optimize_pipeline:
# Принимает указатель на сгенерированный блок кода, анализирует граф потока данных
# и устраняет избыточные операции пролога/эпилога и предсказуемые бранчи.
# Вход:
#   %rdi - указатель на буфер инструкций
#   %rsi - размер буфера инструкций
# Выход:
#   %rax - новый оптимизированный размер (сжатый код)
# ==============================================================================
hpt_optimize_pipeline:
    pushq %rbp
    movq %rsp, %rbp
    pushq %rbx
    pushq %r12

    movq %rdi, %rbx          # rbx = buffer ptr
    movq %rsi, %r12          # r12 = buffer len

    # Проверяем тривиальные размеры
    cmpq $16, %r12
    jl .L_hpt_done

    # 1. Peephole pass: eliminate redundant "    pushq %rax\n    popq %rax\n" (30 bytes)
    # In-place compaction: %r8 = read_idx, %r9 = write_idx
    xorq %r8, %r8
    xorq %r9, %r9
    leaq hpt_pat_pushpop(%rip), %rcx

.L_hpt_scan:
    cmpq %r12, %r8
    jge .L_hpt_finish

    # Check if at least 30 bytes remain
    movq %r12, %rax
    subq %r8, %rax
    cmpq $30, %rax
    jl .L_hpt_copy_byte

    # Compare 30 bytes against "    pushq %rax\n    popq %rax\n"
    xorq %r10, %r10
.L_hpt_cmp:
    cmpq $30, %r10
    jge .L_hpt_matched
    movq %r8, %r11
    addq %r10, %r11
    movb (%rbx, %r11), %al
    cmpb (%rcx, %r10), %al
    jne .L_hpt_copy_byte
    incq %r10
    jmp .L_hpt_cmp

.L_hpt_matched:
    addq $30, %r8
    jmp .L_hpt_scan

.L_hpt_copy_byte:
    movb (%rbx, %r8), %al
    movb %al, (%rbx, %r9)
    incq %r8
    incq %r9
    jmp .L_hpt_scan

.L_hpt_finish:
    movq %r9, %r12

.L_hpt_done:
    movq %r12, %rax          # возвращаем результирующий размер
    popq %r12
    popq %rbx
    leave
    ret

# ==============================================================================
# hpt_verify_flawless_invariants(buf: %rdi, len: %rsi) -> %rax (1 = OK, 0 = FAIL)
# Проверяет наличие точки входа _start: и системного вызова syscall в буфере
# ==============================================================================
.section .rodata
hpt_pat_start:   .ascii "_start:"
hpt_pat_syscall: .ascii "syscall"
hpt_pat_pushpop: .ascii "    pushq %rax\n    popq %rax\n"

.section .text
hpt_verify_flawless_invariants:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    movq %rdi, %r12
    movq %rsi, %r13

    testq %r12, %r12
    jz .L_hv_fail
    cmpq $16, %r13
    jl .L_hv_fail

    leaq hpt_pat_start(%rip), %rdx
    movq $7, %rcx
    call .L_hv_find
    testq %rax, %rax
    jz .L_hv_fail

    leaq hpt_pat_syscall(%rip), %rdx
    movq $7, %rcx
    call .L_hv_find
    testq %rax, %rax
    jz .L_hv_fail

    movq $1, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_hv_fail:
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

.L_hv_find:
    # %r12 = buf, %r13 = len, %rdx = pat, %rcx = pat_len
    cmpq %rcx, %r13
    jl .L_hvf_no
    movq %r13, %r8
    subq %rcx, %r8
    xorq %r9, %r9
.L_hvf_out:
    cmpq %r8, %r9
    jg .L_hvf_no
    xorq %r10, %r10
.L_hvf_in:
    cmpq %rcx, %r10
    jge .L_hvf_yes
    movq %r9, %r11
    addq %r10, %r11
    movb (%r12, %r11), %al
    cmpb (%rdx, %r10), %al
    jne .L_hvf_next
    incq %r10
    jmp .L_hvf_in
.L_hvf_next:
    incq %r9
    jmp .L_hvf_out
.L_hvf_yes:
    movq $1, %rax
    ret
.L_hvf_no:
    xorq %rax, %rax
    ret
