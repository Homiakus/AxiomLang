.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .text
.global mca_arena_create
.global mca_arena_alloc
.global mca_arena_reset
.global mca_arena_destroy

# ==============================================================================
# mca_arena_create(size: %rdi, caps: %rsi) -> %rax (Arena Pointer)
# Uses SYS_MMAP directly without libc/malloc.
# ==============================================================================
mca_arena_create:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    movq %rdi, %r12          # requested size
    movq %rsi, %r13          # capability bitmask

    # Align size to 4096 page boundary
    addq $MCA_HEADER_SIZE, %r12
    addq $4095, %r12
    andq $-4096, %r12

    # mmap(NULL, %r12, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0)
    movq $SYS_MMAP, %rax
    xorq %rdi, %rdi
    movq %r12, %rsi
    movq $3, %rdx            # PROT_READ | PROT_WRITE
    movq $34, %r10           # MAP_PRIVATE | MAP_ANONYMOUS
    movq $-1, %r8
    xorq %r9, %r9
    syscall

    cmpq $0, %rax
    jl mca_create_failed

    # Initialize Arena Header
    movl $MCA_MAGIC, MCA_OFF_MAGIC(%rax)
    movl %r13d, MCA_OFF_CAPS(%rax)
    movq %r12, MCA_OFF_TOTAL_SIZE(%rax)
    movq $MCA_HEADER_SIZE, MCA_OFF_USED_SIZE(%rax)
    movq $0, MCA_OFF_PARENT_PTR(%rax)

    popq %r13
    popq %r12
    leave
    ret

mca_create_failed:
    xorq %rax, %rax
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# mca_arena_alloc(arena: %rdi, size: %rsi) -> %rax (Data Pointer)
# Bump pointer allocation with 8-byte alignment in O(1).
# ==============================================================================
mca_arena_alloc:
    pushq %rbp
    movq %rsp, %rbp

    testq %rdi, %rdi
    jz alloc_oom

    # Verify Magic
    cmpl $MCA_MAGIC, MCA_OFF_MAGIC(%rdi)
    jne alloc_oom

    # Align requested size to 8 bytes
    movq %rsi, %rax
    addq $7, %rax
    andq $-8, %rax
    movq %rax, %rsi

    movq MCA_OFF_USED_SIZE(%rdi), %rcx
    movq %rcx, %rdx
    addq %rsi, %rdx          # new_used = old_used + aligned_size

    cmpq MCA_OFF_TOTAL_SIZE(%rdi), %rdx
    ja alloc_oom             # Out of arena capacity!

    # Commit bump pointer
    movq %rdx, MCA_OFF_USED_SIZE(%rdi)
    leaq (%rdi, %rcx, 1), %rax # return arena_base + old_used

    leave
    ret

alloc_oom:
    xorq %rax, %rax
    leave
    ret

# ==============================================================================
# mca_arena_reset(arena: %rdi)
# Rewinds bump allocator back to 0 (header size) in O(1).
# ==============================================================================
mca_arena_reset:
    testq %rdi, %rdi
    jz reset_done
    cmpl $MCA_MAGIC, MCA_OFF_MAGIC(%rdi)
    jne reset_done
    movq $MCA_HEADER_SIZE, MCA_OFF_USED_SIZE(%rdi)
reset_done:
    ret

# ==============================================================================
# mca_arena_destroy(arena: %rdi)
# Frees entire arena via SYS_MUNMAP.
# ==============================================================================
mca_arena_destroy:
    pushq %rbp
    movq %rsp, %rbp
    testq %rdi, %rdi
    jz destroy_done
    cmpl $MCA_MAGIC, MCA_OFF_MAGIC(%rdi)
    jne destroy_done

    movq MCA_OFF_TOTAL_SIZE(%rdi), %rsi
    movq $SYS_MUNMAP, %rax
    syscall

destroy_done:
    leave
    ret
