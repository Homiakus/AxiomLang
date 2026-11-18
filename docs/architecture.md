# Спецификация и Архитектура AxiomLang (AXL)

AxiomLang — системный язык программирования и freestanding-компилятор для x86_64 Linux без зависимости от `libc` и внешних рантаймов, со статической верификацией полномочий (**Zero-Ambient Authority**) и ареновым управлением памятью **MCA (Mandate Capability Arena)**.

---

## 1. Парадигма Zero-Ambient Authority

Программа на AxiomLang явно декларирует необходимые полномочия (`require [...]`, `cap memory [...]`), а память управляется через регионы MCA с $O(1)$ аллокацией и $O(1)$ сбросом области видимости.

---

## 2. Архитектура Компилятора `axlc_asm` (Pure x86_64 ASM)

1. **CLI-драйвер** ([`src_asm/main.s`](../src_asm/main.s)): разбор аргументов (`-o <binary>`, `-run`), чтение исходника и оркестрация `/usr/bin/as` и `/usr/bin/ld` через `fork` / `execve` / `wait4`.
2. **Лексер** ([`src_asm/core/lexer.s`](../src_asm/core/lexer.s)): zero-copy сканер токенов (`lexer_next_token`), ключевых слов, строк, чисел и операторов.
3. **Парсер и AST** ([`src_asm/parser/ast.s`](../src_asm/parser/ast.s), [`src_asm/include/ast_defs.inc`](../src_asm/include/ast_defs.inc)): построение AST-дерева в MCA-арене (`parser_init`, `parser_parse_program`, `parser_parse_decl`).
4. **Статический верификатор** ([`src_asm/checker/checker.s`](../src_asm/checker/checker.s)): проверка структуры AST, границ (`checker_validate_bounds`) и статических инвариантов (включая деление на ноль).
5. **Отладка DWARF** ([`src_asm/debug/dwarf.s`](../src_asm/debug/dwarf.s)): генерация директив `.file` и `.loc` для отладки в GDB/LLDB.
6. **Кодогенератор и HPT-конвейер** ([`src_asm/compiler.s`](../src_asm/compiler.s), [`src_asm/hpt_opt/hpt_optimizer.s`](../src_asm/hpt_opt/hpt_optimizer.s)): трансляция AST в машинный ассемблер x86_64 с bounds-checked буфером вывода.

---

## 3. Рантайм и Инструментарий

- **[`src_asm/mca_runtime.s`](../src_asm/mca_runtime.s)**: управление аренами памяти через прямые `SYS_MMAP` / `SYS_MUNMAP`.
- **[`src_asm/stdlib/actor_io.s`](../src_asm/stdlib/actor_io.s)**: изолированные почтовые ящики акторов на базе MCA и потоковый вывод `SYS_WRITE`.
- **[`src_asm/axpm/main.s`](../src_asm/axpm/main.s)**: пакетный менеджер `axpm` (`init`, `build`, `run`, `test`).
- **[`src_asm/lsp/main.s`](../src_asm/lsp/main.s)**: freestanding LSP-сервер `axiom_lsp` с диспетчеризацией JSON-RPC.
- **[`src_asm/lean4/lean4_emitter.s`](../src_asm/lean4/lean4_emitter.s)**, **[`src_asm/runtime_heal/runtime_heal.s`](../src_asm/runtime_heal/runtime_heal.s)**, **[`src_asm/actor_sheaf/actor_sheaf.s`](../src_asm/actor_sheaf/actor_sheaf.s)**: модули формальных сертификатов Lean 4 и восстановления инвариантов.
