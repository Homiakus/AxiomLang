; ModuleID = 'AxiomBareMetalKernel'
source_filename = "kernel.axl"
target datalayout = "e-m:e-p:32:32-i64:64-n32-S128"
target triple = "riscv32imc-unknown-none-elf"

; Freestanding Kernel Intrinsics
declare void @llvm.x86.sse.pause() nounwind
declare void @llvm.donothing() nounwind readnone

; Hardware Video Memory Mapped I/O Framebuffer Routines (Zero OS / Freestanding)
; Writes a 32-bit pixel directly to physical linear framebuffer address: addr + (y * pitch) + (x * 4)
define void @baremetal_fb_draw_pixel(i64 %fb_base, i64 %pitch, i64 %x, i64 %y, i32 %color) nounwind {
entry:
  %row_off = mul i64 %y, %pitch
  %col_off = shl i64 %x, 2
  %tot_off = add i64 %row_off, %col_off
  %pix_addr = add i64 %fb_base, %tot_off
  %ptr = inttoptr i64 %pix_addr to i32*
  store i32 %color, i32* %ptr, align 4
  ret void
}

; Writes ASCII character to standard VGA Text Buffer 0xB8000 (Ring 0)
define void @baremetal_vga_write_char(i64 %col, i64 %row, i8 %ch, i8 %color_attr) nounwind {
entry:
  %row_bytes = mul i64 %row, 160
  %col_bytes = shl i64 %col, 1
  %off = add i64 %row_bytes, %col_bytes
  %vga_addr = add i64 753664, %off ; 0xB8000 = 753664
  %ch_ptr = inttoptr i64 %vga_addr to i8*
  store i8 %ch, i8* %ch_ptr, align 1
  %attr_addr = add i64 %vga_addr, 1
  %attr_ptr = inttoptr i64 %attr_addr to i8*
  store i8 %color_attr, i8* %attr_ptr, align 1
  ret void
}

%MCUSystemState = type { i64, i64, i8 }

define i64 @configure_gpio(i64 %port_base, i64 %pin_mask) nounwind {
entry:
  %t1 = load i64, i64* %port_base, align 8
  %t2 = load i64, i64* %pin_mask, align 8
  %ptr3 = inttoptr i64 %t1 to i32*
  store volatile i32 %t2, i32* %ptr3, align 4
  %t4 = load i64, i64* %port_base, align 8
  %ptr5 = inttoptr i64 %t4 to i32*
  %t6 = load volatile i32, i32* %ptr5, align 4
  ret i64 %t6
}

define i64 @set_actuator_pin(i64 %port_base, i64 %state) nounwind {
entry:
  %t7 = load i64, i64* %port_base, align 8
  %t8 = add i64 %t7, 24
  %t9 = load i64, i64* %state, align 8
  %ptr10 = inttoptr i64 %t8 to i32*
  store volatile i32 %t9, i32* %ptr10, align 4
  ret i64 1
}

; Kernel Entry Point for Bootloader (Multiboot / Limine / UEFI)
define void @_start() noreturn nounwind section ".text.boot" {
entry:
  ; Phase Intent: Инициализация тактирования и защищенной периферии микроконтроллера
  call void asm sideeffect "cpsid i", "~{memory}"() nounwind
  %t11 = call i64 @configure_gpio(i64 1073872896, i64 65535)
  %gpio_status = alloca i64, align 8
  store i64 %t11, i64* %gpio_status, align 8
  %ptr12 = inttoptr i64 1073811456 to i32*
  store volatile i32 115200, i32* %ptr12, align 4
  call void asm sideeffect "cpsie i", "~{memory}"() nounwind
  ; Phase Intent: Рабочий цикл защитной автоматики (Safety Interlock Loop)
  %state = alloca i64, align 8
  store i64 0, i64* %state, align 8
  %t13 = call i64 @set_actuator_pin(i64 1073872896, i64 1)
  ; Bare-metal halt loop: cli; hlt
  call void asm sideeffect "cli; hlt", "~{memory}"() nounwind
  unreachable
}
