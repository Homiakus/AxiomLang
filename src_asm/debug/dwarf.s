.include "src_asm/include/axiom_defs.inc"

.section .data
# DWARF Debug Directives template
debug_file_prefix:
    .ascii "\n.file 1 \""
debug_file_prefix_len = . - debug_file_prefix

debug_file_suffix:
    .ascii "\"\n"
debug_file_suffix_len = . - debug_file_suffix

debug_loc_prefix:
    .ascii "    .loc 1 "
debug_loc_prefix_len = . - debug_loc_prefix

debug_loc_suffix:
    .ascii " 0\n"
debug_loc_suffix_len = . - debug_loc_suffix

.section .text
.global dwarf_emit_file
.global dwarf_emit_loc

# ==============================================================================
# dwarf_emit_file(filename_ptr: %rdi, filename_len: %rsi)
# Emits .file 1 "<filename>" to generated assembly for GDB/LLDB line tables
# ==============================================================================
dwarf_emit_file:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13

    movq %rdi, %r12
    movq %rsi, %r13

    # Emit .file 1 "
    leaq debug_file_prefix(%rip), %rsi
    movq $debug_file_prefix_len, %rcx
    call append_to_output

    # Emit path
    movq %r12, %rsi
    movq %r13, %rcx
    call append_to_output

    # Emit "\n
    leaq debug_file_suffix(%rip), %rsi
    movq $debug_file_suffix_len, %rcx
    call append_to_output

    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# dwarf_emit_loc(line_number: %rdi)
# Emits .loc 1 <line_number> 0 to output stream
# ==============================================================================
dwarf_emit_loc:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12

    movq %rdi, %r12

    leaq debug_loc_prefix(%rip), %rsi
    movq $debug_loc_prefix_len, %rcx
    call append_to_output

    # Convert line number to ascii
    movq %r12, %rax
    call emit_integer_ascii

    leaq debug_loc_suffix(%rip), %rsi
    movq $debug_loc_suffix_len, %rcx
    call append_to_output

    popq %r12
    leave
    ret

# Helper: emit integer %rax in ASCII to output stream
emit_integer_ascii:
    pushq %rbx
    pushq %rdx
    subq $32, %rsp

    leaq 31(%rsp), %rsi
    movb $0, (%rsi)
    movq $0, %rcx          # char count

    testq %rax, %rax
    jnz int_conv_loop
    # Handle zero
    decq %rsi
    movb $'0', (%rsi)
    incq %rcx
    jmp int_conv_done

int_conv_loop:
    testq %rax, %rax
    jz int_conv_done
    xorq %rdx, %rdx
    movq $10, %rbx
    divq %rbx              # rax = rax / 10, rdx = remainder
    addb $'0', %dl
    decq %rsi
    movb %dl, (%rsi)
    incq %rcx
    jmp int_conv_loop

int_conv_done:
    # Append converted string from %rsi with length %rcx
    call append_to_output
    addq $32, %rsp
    popq %rdx
    popq %rbx
    ret
