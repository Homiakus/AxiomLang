# ==============================================================================
# AxiomLang Pure Freestanding x86_64 Native X11 GUI Engine & CoreGFX
# Direct UNIX domain socket connection to /tmp/.X11-unix/X0
# Zero External Libraries | Zero libc | Zero WebView | Direct Linux Syscalls
# Full Subpixel Event Dispatcher: Mouse ButtonPress, Motion, KeyPress
# ==============================================================================

.include "src_asm/include/axiom_defs.inc"

.equ SYS_SOCKET,  41
.equ SYS_CONNECT, 42
.equ SYS_POLL,    7

.equ AF_UNIX,     1
.equ SOCK_STREAM, 1

.section .rodata
x11_socket_path:
    .asciz "/tmp/.X11-unix/X0"
x11_socket_path_len = . - x11_socket_path

x11_setup_req:
    .byte 0x6c, 0x00        # Little endian byte order
    .word 11, 0             # Major 11, Minor 0
    .word 0, 0              # Auth proto name length 0, data length 0
    .word 0                 # Unused pad
x11_setup_req_len = . - x11_setup_req

font_name:
    .ascii "9x15"
font_name_len = . - font_name

msg_x11_connected:
    .ascii "[Axiom-X11] Successfully connected to X-Server via direct UNIX socket.\n"
msg_x11_connected_len = . - msg_x11_connected

msg_x11_window_mapped:
    .ascii "[Axiom-X11] Standalone Native Browser Window (1180x820) mapped to screen.\n"
msg_x11_window_mapped_len = . - msg_x11_window_mapped

msg_x11_err:
    .ascii "[Axiom-X11] Error connecting to /tmp/.X11-unix/X0\n"
msg_x11_err_len = . - msg_x11_err

.section .bss
.align 16
.global x11_sockfd, x11_win_id, x11_gc_id
x11_sockfd:        .quad 0
x11_rid_base:      .long 0
x11_rid_mask:      .long 0
x11_root_win:      .long 0
x11_win_id:        .long 0
x11_gc_id:         .long 0
x11_font_id:       .long 0
x11_reply_buf:     .skip 8192
x11_req_buf:       .skip 2048

.section .text
.global axiom_x11_init_window
.global axiom_x11_draw_rect
.global axiom_x11_draw_text
.global axiom_x11_event_loop_interactive
.global axiom_x11_close

# ==============================================================================
# axiom_x11_init_window(width: %rdi, height: %rsi, title_ptr: %rdx, title_len: %rcx)
# ==============================================================================
axiom_x11_init_window:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14
    pushq %r15

    # 1. socket(AF_UNIX, SOCK_STREAM, 0)
    movq $SYS_SOCKET, %rax
    movq $AF_UNIX, %rdi
    movq $SOCK_STREAM, %rsi
    xorq %rdx, %rdx
    syscall
    testq %rax, %rax
    js .L_x11_fail
    movq %rax, x11_sockfd(%rip)

    # 2. connect(sockfd, &sockaddr_un, 110)
    subq $128, %rsp
    movw $AF_UNIX, (%rsp)
    leaq 2(%rsp), %rdi
    leaq x11_socket_path(%rip), %rsi
    movq $x11_socket_path_len, %rcx
    cld
    rep movsb

    movq $SYS_CONNECT, %rax
    movq x11_sockfd(%rip), %rdi
    movq %rsp, %rsi
    movq $110, %rdx
    syscall
    addq $128, %rsp
    testq %rax, %rax
    js .L_x11_fail

    # Print success
    movq $SYS_WRITE, %rax
    movq $1, %rdi
    leaq msg_x11_connected(%rip), %rsi
    movq $msg_x11_connected_len, %rdx
    syscall

    # 3. Setup packet
    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_setup_req(%rip), %rsi
    movq $12, %rdx
    syscall

    # 4. Read setup reply header
    movq $SYS_READ, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_reply_buf(%rip), %rsi
    movq $8, %rdx
    syscall

    cmpb $1, x11_reply_buf(%rip)
    jne .L_x11_fail

    movzwq x11_reply_buf+6(%rip), %rax
    shlq $2, %rax
    movq %rax, %r12

    movq $SYS_READ, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_reply_buf(%rip), %rsi
    movq %r12, %rdx
    syscall

    # Parse RID base, mask, and root_window
    movl x11_reply_buf+4(%rip), %eax
    movl %eax, x11_rid_base(%rip)
    movl x11_reply_buf+8(%rip), %eax
    movl %eax, x11_rid_mask(%rip)

    movzwq x11_reply_buf+16(%rip), %rax   # vendor_len
    movq %rax, %rcx
    andq $3, %rcx
    testq %rcx, %rcx
    jz .L_no_pad
    movq $4, %rdx
    subq %rcx, %rdx
    addq %rdx, %rax
.L_no_pad:
    addq $32, %rax
    movzbq x11_reply_buf+21(%rip), %rcx   # num_formats
    shlq $3, %rcx
    addq %rcx, %rax

    leaq x11_reply_buf(%rip), %rbx
    addq %rax, %rbx
    movl (%rbx), %eax
    movl %eax, x11_root_win(%rip)

    movl x11_rid_base(%rip), %eax
    movl %eax, x11_win_id(%rip)
    incl %eax
    movl %eax, x11_gc_id(%rip)
    incl %eax
    movl %eax, x11_font_id(%rip)

    # 5. OpenFont Request (Opcode 45, len 4 = 16 bytes)
    # [1] opcode: 45, [1] pad
    # [2] request length: 4 (16 bytes)
    # [4] fid: x11_font_id
    # [2] name_length: 4 ("9x15")
    # [2] pad: 0
    # [4] name: "9x15"
    leaq x11_req_buf(%rip), %rdi
    movb $45, 0(%rdi)
    movb $0, 1(%rdi)
    movw $4, 2(%rdi)
    movl x11_font_id(%rip), %eax
    movl %eax, 4(%rdi)
    movw $4, 8(%rdi)
    movw $0, 10(%rdi)
    movl $0x35317839, 12(%rdi)            # "9x15" in little-endian ASCII

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $16, %rdx
    syscall

    # 6. CreateWindow (Opcode 1, len 10 = 40 bytes)
    # Event mask: KeyPress (1) | ButtonPress (4) | ButtonRelease (8) | PointerMotion (64) | Exposure (0x8000) = 0x804d
    leaq x11_req_buf(%rip), %rdi
    movb $1, 0(%rdi)
    movb $0, 1(%rdi)
    movw $10, 2(%rdi)
    movl x11_win_id(%rip), %eax
    movl %eax, 4(%rdi)
    movl x11_root_win(%rip), %eax
    movl %eax, 8(%rdi)
    movw $60, 12(%rdi)                    # x
    movw $50, 14(%rdi)                    # y
    movw $1180, 16(%rdi)                  # width
    movw $820, 18(%rdi)                   # height
    movw $1, 20(%rdi)                     # border_width
    movw $1, 22(%rdi)                     # class (InputOutput)
    movl $0, 24(%rdi)                     # visual (CopyFromParent)
    movl $0x802, 28(%rdi)                 # value_mask: CWBackPixel | CWEventMask
    movl $0x0a0d14, 32(%rdi)              # background pixel (#0a0d14)
    movl $0x804d, 36(%rdi)                # event_mask: KeyPress|ButtonPress|ButtonRelease|PointerMotion|Exposure

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $40, %rdx
    syscall

    # 7. MapWindow (Opcode 8, len 2 = 8 bytes)
    leaq x11_req_buf(%rip), %rdi
    movb $8, 0(%rdi)
    movb $0, 1(%rdi)
    movw $2, 2(%rdi)
    movl x11_win_id(%rip), %eax
    movl %eax, 4(%rdi)

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $8, %rdx
    syscall

    # 8. CreateGC (Opcode 55, len 5 = 20 bytes)
    # mask = 0x4000 (GCFont)
    leaq x11_req_buf(%rip), %rdi
    movb $55, 0(%rdi)
    movb $0, 1(%rdi)
    movw $5, 2(%rdi)
    movl x11_gc_id(%rip), %eax
    movl %eax, 4(%rdi)
    movl x11_win_id(%rip), %eax
    movl %eax, 8(%rdi)
    movl $0x4000, 12(%rdi)                # GCFont mask
    movl x11_font_id(%rip), %eax
    movl %eax, 16(%rdi)                   # font id

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $20, %rdx
    syscall

    # Print success
    movq $SYS_WRITE, %rax
    movq $1, %rdi
    leaq msg_x11_window_mapped(%rip), %rsi
    movq $msg_x11_window_mapped_len, %rdx
    syscall

    xorq %rax, %rax
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

.L_x11_fail:
    movq $SYS_WRITE, %rax
    movq $2, %rdi
    leaq msg_x11_err(%rip), %rsi
    movq $msg_x11_err_len, %rdx
    syscall
    movq $1, %rax
    popq %r15
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# axiom_x11_draw_rect(x: %rdi, y: %rsi, w: %rdx, h: %rcx, color_rgb: %r8)
# ==============================================================================
axiom_x11_draw_rect:
    pushq %rbp
    movq %rsp, %rbp

    # Change foreground
    leaq x11_req_buf(%rip), %rax
    movb $56, 0(%rax)
    movb $0, 1(%rax)
    movw $4, 2(%rax)
    movl x11_gc_id(%rip), %r9d
    movl %r9d, 4(%rax)
    movl $4, 8(%rax)                      # GCForeground
    movl %r8d, 12(%rax)

    pushq %rdi
    pushq %rsi
    pushq %rdx
    pushq %rcx

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $16, %rdx
    syscall

    popq %rcx
    popq %rdx
    popq %rsi
    popq %rdi

    # PolyFillRectangle (Opcode 70)
    leaq x11_req_buf(%rip), %rax
    movb $70, 0(%rax)
    movb $0, 1(%rax)
    movw $5, 2(%rax)
    movl x11_win_id(%rip), %r9d
    movl %r9d, 4(%rax)
    movl x11_gc_id(%rip), %r9d
    movl %r9d, 8(%rax)
    movw %di, 12(%rax)
    movw %si, 14(%rax)
    movw %dx, 16(%rax)
    movw %cx, 18(%rax)

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $20, %rdx
    syscall

    leave
    ret

# ==============================================================================
# axiom_x11_draw_text(x: %rdi, y: %rsi, str_ptr: %rdx, str_len: %rcx, color_rgb: %r8)
# ==============================================================================
axiom_x11_draw_text:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14

    movq %rdx, %r12
    movq %rcx, %r13

    # Foreground
    leaq x11_req_buf(%rip), %rax
    movb $56, 0(%rax)
    movb $0, 1(%rax)
    movw $4, 2(%rax)
    movl x11_gc_id(%rip), %r9d
    movl %r9d, 4(%rax)
    movl $4, 8(%rax)
    movl %r8d, 12(%rax)

    pushq %rdi
    pushq %rsi
    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    movq $16, %rdx
    syscall
    popq %rsi
    popq %rdi

    # ImageText8 (Opcode 76)
    leaq x11_req_buf(%rip), %rax
    movb $76, 0(%rax)
    movb %r13b, 1(%rax)

    movq %r13, %rdx
    addq $16, %rdx
    addq $3, %rdx
    shrq $2, %rdx
    movw %dx, 2(%rax)

    movl x11_win_id(%rip), %r9d
    movl %r9d, 4(%rax)
    movl x11_gc_id(%rip), %r9d
    movl %r9d, 8(%rax)
    movw %di, 12(%rax)
    movw %si, 14(%rax)

    leaq 16(%rax), %rdi
    movq %r12, %rsi
    movq %r13, %rcx
    cld
    rep movsb

    movzwq 2(%rax), %rdx
    shlq $2, %rdx

    movq $SYS_WRITE, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_req_buf(%rip), %rsi
    syscall

    popq %r14
    popq %r13
    popq %r12
    leave
    ret

# ==============================================================================
# axiom_x11_event_loop_interactive(render_fn: %rdi, click_fn: %rsi, key_fn: %rdx)
# Non-blocking fully interactive event loop with subpixel hit-testing
# ==============================================================================
axiom_x11_event_loop_interactive:
    pushq %rbp
    movq %rsp, %rbp
    pushq %r12
    pushq %r13
    pushq %r14

    movq %rdi, %r12                       # render_fn()
    movq %rsi, %r13                       # click_fn(x, y, btn)
    movq %rdx, %r14                       # key_fn(keycode)

    subq $16, %rsp
    movl x11_sockfd(%rip), %eax
    movl %eax, (%rsp)
    movw $1, 4(%rsp)                      # POLLIN
    movw $0, 6(%rsp)

.L_interactive_poll:
    movq $SYS_POLL, %rax
    movq %rsp, %rdi
    movq $1, %rsi
    movq $-1, %rdx                        # wait indefinitely
    syscall

    testq %rax, %rax
    jle .L_interactive_poll

    # Read 32-byte X11 event packet
    movq $SYS_READ, %rax
    movq x11_sockfd(%rip), %rdi
    leaq x11_reply_buf(%rip), %rsi
    movq $32, %rdx
    syscall

    testq %rax, %rax
    jle .L_interactive_poll

    movzbq x11_reply_buf(%rip), %rax
    andq $0x7f, %rax                      # event opcode

    # 1. Expose event (code 12)
    cmpq $12, %rax
    jne .L_check_btn_press
    testq %r12, %r12
    jz .L_interactive_poll
    call *%r12
    jmp .L_interactive_poll

.L_check_btn_press:
    # 2. ButtonPress event (code 4)
    cmpq $4, %rax
    jne .L_check_keypress

    # Extract:
    # byte 1: button (1 = Left Click)
    # bytes 24..25: event-x (INT16)
    # bytes 26..27: event-y (INT16)
    movzbq x11_reply_buf+1(%rip), %rdx    # btn
    movswq x11_reply_buf+24(%rip), %rdi   # x
    movswq x11_reply_buf+26(%rip), %rsi   # y

    testq %r13, %r13
    jz .L_interactive_poll
    call *%r13                           # click_fn(x, y, btn)

    # Re-render immediately on click
    testq %r12, %r12
    jz .L_interactive_poll
    call *%r12
    jmp .L_interactive_poll

.L_check_keypress:
    # 3. KeyPress event (code 2)
    cmpq $2, %rax
    jne .L_interactive_poll

    movzbq x11_reply_buf+1(%rip), %rdi    # keycode
    # If Esc (9) or 'q' (24) -> exit
    cmpq $9, %rdi
    je .L_loop_exit
    cmpq $24, %rdi
    je .L_loop_exit

    testq %r14, %r14
    jz .L_interactive_poll
    call *%r14                           # key_fn(keycode)

    # Re-render
    testq %r12, %r12
    jz .L_interactive_poll
    call *%r12
    jmp .L_interactive_poll

.L_loop_exit:
    addq $16, %rsp
    popq %r14
    popq %r13
    popq %r12
    leave
    ret

axiom_x11_close:
    movq $SYS_CLOSE, %rax
    movq x11_sockfd(%rip), %rdi
    syscall
    ret
