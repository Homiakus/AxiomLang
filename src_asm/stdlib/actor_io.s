.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .text
.global stdlib_actor_spawn
.global stdlib_actor_send
.global stdlib_io_print_str

# ==============================================================================
# stdlib_actor_spawn(entry_fn: %rdi, arena_size: %rsi) -> %rax (Actor Token ID)
# Zero-cost lightweight pure ASM actor creation
# ==============================================================================
stdlib_actor_spawn:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12

    movq %rdi, %r12
    # Create isolated MCA arena for actor mailbox and scratchpad
    movq %rsi, %rdi
    movq $7, %rsi            # CAP_MEM_READ | CAP_MEM_WRITE | CAP_MEM_EXEC
    call mca_arena_create

    # Actor token equals arena base pointer
    popq %r12
    leave
    ret

# ==============================================================================
# stdlib_actor_send(actor_token: %rdi, msg_ptr: %rsi, msg_len: %rdx) -> %rax
# Direct message copying into actor's mailbox
# ==============================================================================
stdlib_actor_send:
    pushq %rbp
    movq %rsp, %rbp

    # Validate destination actor token
    testq %rdi, %rdi
    jz send_err

    # Allocation inside recipient actor arena
    pushq %rsi
    pushq %rdx
    pushq %rdi
    movq %rdx, %rsi
    call mca_arena_alloc
    popq %rdi
    popq %rdx
    popq %rsi

    testq %rax, %rax
    jz send_err

    # Copy message
    pushq %rdi
    movq %rax, %rdi          # dest
    movq %rdx, %rcx          # count
    cld
    rep movsb
    popq %rdi

    xorq %rax, %rax          # success
    leave
    ret

send_err:
    movq $1, %rax
    leave
    ret

# ==============================================================================
# stdlib_io_print_str(str_ptr: %rdi, str_len: %rsi)
# Streaming syscall print
# ==============================================================================
stdlib_io_print_str:
    movq $SYS_WRITE, %rax
    movq %rsi, %rdx
    movq %rdi, %rsi
    movq $STDOUT_FILENO, %rdi
    syscall
    ret
