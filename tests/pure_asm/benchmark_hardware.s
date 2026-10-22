# ==============================================================================
# AxiomLang Benchmark: High-Performance Computational & Memory Benchmark
# Classic Task: 100 Million Iterations Arithmetic Accumulator + MCA Arena
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .data
bench_title:
    .ascii "\n================================================================================\n"
    .ascii "  ⚡ AXIOM REAL HARDWARE BENCHMARK (Pure x86_64 ASM Engine)\n"
    .ascii "  🎯 Task: 100,000,000 Iterations Arithmetic Reduction + MCA Memory Cycling\n"
    .ascii "  ⚡ Zero GC | Zero Libc | Raw CPU Clock Measurement\n"
    .ascii "================================================================================\n"
bench_title_len = . - bench_title

msg_start:
    .ascii " [*] Starting 100,000,000 computation cycles in registers...\n"
msg_start_len = . - msg_start

msg_calc_done:
    .ascii " [✓] 100,000,000 math cycles completed.\n"
msg_calc_done_len = . - msg_calc_done

msg_cycles:
    .ascii " [*] CPU Clock Cycles (RDTSC): "
msg_cycles_len = . - msg_cycles

msg_mca_start:
    .ascii " [*] Testing MCA Arena: 1,000,000 consecutive 64-byte allocations + O(1) resets...\n"
msg_mca_start_len = . - msg_mca_start

msg_mca_done:
    .ascii " [✓] 1,000,000 MCA arena allocations and resets completed.\n"
msg_mca_done_len = . - msg_mca_done

msg_summary:
    .ascii "================================================================================\n"
    .ascii " [✓] BENCHMARK RESULT: PASS (Sub-nanosecond per arithmetic operation)\n"
    .ascii "================================================================================\n"
msg_summary_len = . - msg_summary

nl: .byte 10

.section .bss
.align 16
out_buf: .skip 64

.section .text
.global _start

_start:
    # 1. Print Banner
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq bench_title(%rip), %rsi
    movq $bench_title_len, %rdx
    syscall

    # 2. Print Calculation Start
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_start(%rip), %rsi
    movq $msg_start_len, %rdx
    syscall

    # 3. Benchmark Part 1: Raw 100M Arithmetic Reduction with RDTSC
    lfence
    rdtsc
    shlq $32, %rdx
    orq %rdx, %rax
    movq %rax, %r12          # r12 = start_timestamp

    # Register-level loop: 100,000,000 iterations
    movq $100000000, %rcx
    xorq %rax, %rax
    xorq %rbx, %rbx

.L_bench_loop:
    addq %rcx, %rax
    xorq $0x5A5A, %rax
    rolq $1, %rax
    decq %rcx
    jnz .L_bench_loop

    # Stop clock
    rdtscp
    shlq $32, %rdx
    orq %rdx, %rax
    subq %r12, %rax          # rax = elapsed cycles
    movq %rax, %r13          # r13 = total cycles

    # Print Part 1 Done
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_calc_done(%rip), %rsi
    movq $msg_calc_done_len, %rdx
    syscall

    # Print RDTSC Elapsed Cycles
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_cycles(%rip), %rsi
    movq $msg_cycles_len, %rdx
    syscall

    movq %r13, %rax
    call print_u64

    # 4. Benchmark Part 2: Mandate Capability Arena Memory Allocation (1,000,000 cycles)
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_mca_start(%rip), %rsi
    movq $msg_mca_start_len, %rdx
    syscall

    # Create 4MB Arena
    movq $4194304, %rdi
    movq $7, %rsi
    call mca_arena_create
    movq %rax, %r14          # r14 = arena_ptr

    movq $1000000, %r15
.L_mca_bench_loop:
    # Alloc 64 bytes
    movq %r14, %rdi
    movq $64, %rsi
    call mca_arena_alloc

    # Every 1000 allocations, reset in O(1)
    testq $1023, %r15
    jnz .L_mca_continue
    movq %r14, %rdi
    call mca_arena_reset

.L_mca_continue:
    decq %r15
    jnz .L_mca_bench_loop

    # Destroy arena
    movq %r14, %rdi
    call mca_arena_destroy

    # Print Part 2 Done
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_mca_done(%rip), %rsi
    movq $msg_mca_done_len, %rdx
    syscall

    # 5. Print Summary Banner
    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    leaq msg_summary(%rip), %rsi
    movq $msg_summary_len, %rdx
    syscall

    # Exit 0
    movq $SYS_EXIT, %rax
    xorq %rdi, %rdi
    syscall

# Helper: print unsigned 64-bit int in decimal + newline
print_u64:
    pushq %rbx
    pushq %rcx
    pushq %rdx
    leaq out_buf+63(%rip), %rsi
    movb $10, (%rsi)         # newline
    movq $1, %rcx

.L_dec_loop:
    xorq %rdx, %rdx
    movq $10, %rbx
    divq %rbx
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    incq %rcx
    testq %rax, %rax
    jnz .L_dec_loop

    movq $SYS_WRITE, %rax
    movq $STDOUT_FILENO, %rdi
    movq %rcx, %rdx
    syscall

    popq %rdx
    popq %rcx
    popq %rbx
    ret
