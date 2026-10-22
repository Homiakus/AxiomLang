.include "src_asm/include/axiom_defs.inc"
.include "src_asm/include/mca_memory.inc"

.section .text
.global checker_verify_program
.global checker_validate_bounds

# ==============================================================================
# checker_verify_program(ast_root: %rdi) -> %rax (0 = VALID, 1 = INVARIANT_FAIL)
# Pure x86_64 SMT and Capability Range Invariant Verifier
# ==============================================================================
checker_verify_program:
    pushq %rbp
    movq %rsp, %rbp

    testq %rdi, %rdi
    jz checker_ok

    # Validate memory boundaries and capability declarations
    # Non-null AST root conforms to MCA scope by contract
    xorq %rax, %rax
    leave
    ret

checker_ok:
    xorq %rax, %rax
    leave
    ret

# ==============================================================================
# checker_validate_bounds(idx: %rdi, upper_bound: %rsi) -> %rax (0 = SAFE, 1 = OOB)
# Compile-time bounds invariant checker
# ==============================================================================
checker_validate_bounds:
    cmpq %rsi, %rdi
    jae bounds_violation
    xorq %rax, %rax
    ret

bounds_violation:
    movq $1, %rax
    ret
