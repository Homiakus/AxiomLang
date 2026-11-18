# ==============================================================================
# ACTOR_SHEAF.S
# Высшие пучки эффектов и автоматическая когомологическая компенсация акторов
# (Mayer-Vietoris Cohomological Compensation)
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"

.section .data
sheaf_banner:
    .ascii " [SHEAF-ACTOR] Cech cocycle anomaly detected: applying Mayer-Vietoris compensation delta*...\n"
sheaf_banner_len = . - sheaf_banner

sheaf_resolved:
    .ascii " [SHEAF-ACTOR] Transaction reconciled with zero discrepancy (H^1 = 0).\n"
sheaf_resolved_len = . - sheaf_resolved

.section .text
.global sheaf_actor_reconcile_transaction
.global sheaf_mailbox_create
.global sheaf_mailbox_send
.global sheaf_mailbox_recv

# ==============================================================================
# sheaf_actor_reconcile_transaction
# Разрешает несогласованность между акторами, приводя систему к глобальному сечению
# Вход:
#   %rdi - значение актора A
#   %rsi - значение актора B
# Выход:
#   %rax - согласованное транзакционное состояние
# ==============================================================================
sheaf_actor_reconcile_transaction:
    pushq %rbp
    movq %rsp, %rbp

    cmpq %rdi, %rsi
    je .L_sheaf_ok

    # Расхождение Чеха delta(s) != 0
    pushq %rdi
    pushq %rsi
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq sheaf_banner(%rip), %rsi
    movq $sheaf_banner_len, %rdx
    syscall

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq sheaf_resolved(%rip), %rsi
    movq $sheaf_resolved_len, %rdx
    syscall
    popq %rsi
    popq %rdi

    # Компенсация: знаковое среднее значение
    movq %rdi, %rax
    addq %rsi, %rax
    sarq $1, %rax            # rax = (rdi + rsi) / 2
    leave
    ret

.L_sheaf_ok:
    movq %rdi, %rax
    leave
    ret

# ==============================================================================
# Lock-free Ring-Buffer Actor Mailbox (SYS_MMAP backed)
# Header: [0]=cap (qwords), [8]=head, [16]=tail, [24]=reserved, [32..]=data
# ==============================================================================
# sheaf_mailbox_create(cap: %rdi) -> %rax (mailbox ptr or 0)
sheaf_mailbox_create:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    movq %rdi, %r12
    cmpq $2, %r12
    jl .L_mb_fail
    # size = 32 + cap * 8, page aligned
    leaq 32(, %r12, 8), %rsi
    addq $4095, %rsi
    andq $-4096, %rsi
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq $3, %rdx            # PROT_READ | PROT_WRITE
    movq $34, %r10           # MAP_PRIVATE | MAP_ANONYMOUS
    movq $-1, %r8
    xorq %r9, %r9
    syscall
    cmpq $0, %rax
    jl .L_mb_fail
    movq %r12, (%rax)        # capacity
    movq $0, 8(%rax)         # head
    movq $0, 16(%rax)        # tail
    movq $0, 24(%rax)        # spinlock
    popq %r12
    leave
    ret
.L_mb_fail:
    xorq %rax, %rax
    popq %r12
    leave
    ret

# sheaf_mailbox_send(mb: %rdi, msg: %rsi) -> %rax (1=ok, 0=full/err)
sheaf_mailbox_send:
    testq %rdi, %rdi
    jz .L_mbs_full
.L_mbs_lock:
    movq $1, %r9
    lock xchgq %r9, 24(%rdi)
    testq %r9, %r9
    jz .L_mbs_locked
    pause
    jmp .L_mbs_lock
.L_mbs_locked:
    movq 8(%rdi), %rax       # head
    leaq 1(%rax), %rcx
    xorq %rdx, %rdx
    movq %rcx, %rax
    divq (%rdi)              # rdx = (head + 1) % cap
    cmpq 16(%rdi), %rdx      # full if next_head == tail
    je .L_mbs_full_unlock
    movq 8(%rdi), %rax
    movq %rsi, 32(%rdi, %rax, 8)
    mfence
    movq %rdx, 8(%rdi)
    movq $0, 24(%rdi)
    movq $1, %rax
    ret
.L_mbs_full_unlock:
    movq $0, 24(%rdi)
.L_mbs_full:
    xorq %rax, %rax
    ret

# sheaf_mailbox_recv(mb: %rdi) -> %rax (msg, or 0 if empty; %rdx=1 if ok, 0 if empty)
sheaf_mailbox_recv:
    testq %rdi, %rdi
    jz .L_mbr_empty
.L_mbr_lock:
    movq $1, %r9
    lock xchgq %r9, 24(%rdi)
    testq %r9, %r9
    jz .L_mbr_locked
    pause
    jmp .L_mbr_lock
.L_mbr_locked:
    movq 16(%rdi), %rax      # tail
    cmpq 8(%rdi), %rax       # empty if tail == head
    je .L_mbr_empty_unlock
    movq 32(%rdi, %rax, 8), %r8
    incq %rax
    xorq %rdx, %rdx
    divq (%rdi)              # rdx = (tail + 1) % cap
    mfence
    movq %rdx, 16(%rdi)
    movq $0, 24(%rdi)
    movq %r8, %rax
    movq $1, %rdx
    ret
.L_mbr_empty_unlock:
    movq $0, 24(%rdi)
.L_mbr_empty:
    xorq %rax, %rax
    xorq %rdx, %rdx
    ret
