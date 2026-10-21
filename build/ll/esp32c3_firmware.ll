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

%ESP32Health = type { i64, i64, i64 }

define i64 @esp32_gpio_enable(i64 %pin_mask) nounwind {
entry:
  %t1 = load i64, i64* %pin_mask, align 8
  %ptr2 = inttoptr i64 1610629156 to i32*
  store volatile i32 %t1, i32* %ptr2, align 4
  ret i64 1
}

define i64 @esp32_led_on(i64 %pin_mask) nounwind {
entry:
  %t3 = load i64, i64* %pin_mask, align 8
  %ptr4 = inttoptr i64 1610629128 to i32*
  store volatile i32 %t3, i32* %ptr4, align 4
  ret i64 1
}

define i64 @esp32_led_off(i64 %pin_mask) nounwind {
entry:
  %t5 = load i64, i64* %pin_mask, align 8
  %ptr6 = inttoptr i64 1610629132 to i32*
  store volatile i32 %t5, i32* %ptr6, align 4
  ret i64 0
}

; Kernel Entry Point for Bootloader (Multiboot / Limine / UEFI)
define void @_start() noreturn nounwind section ".text.boot" {
entry:
  ; Phase Intent: Инициализация подсистемы ввода/вывода ESP32-C3
  %t7 = call i64 @esp32_gpio_enable(i64 256)
  %enabled = alloca i64, align 8
  store i64 %t7, i64* %enabled, align 8
  ; Phase Intent: Цикл защитной телеметрии и мигания LED
  %health = alloca i64, align 8
  store i64 0, i64* %health, align 8
  %t8 = call i64 @esp32_led_on(i64 256)
  ; Bare-metal halt loop: cli; hlt
  call void asm sideeffect "cli; hlt", "~{memory}"() nounwind
  unreachable
}
