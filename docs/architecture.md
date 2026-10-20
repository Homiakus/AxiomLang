# Спецификация и Архитектура AxiomLang (AXL)

AxiomLang — язык программирования и компилятор со статической верификацией полномочий (**Zero-Ambient Authority**), защитой от обратной разработки на основе криптографических решеток (**Ring-LWE Obfuscation**) и аппаратной колоночной обработкой данных.

---

## 1. Парадигма Zero-Ambient Authority

В отличие от традиционных сред исполнения, программа на AxiomLang не имеет неявных системных прав:
- Каждая операция ввода/вывода (консоль, диск, сеть, GPU) требует явного объявления мандата `require cap[...]`.
- Статический анализатор [`pkg/checker/checker.go`](file:///home/den/AxiomLang/pkg/checker/checker.go) проверяет зависимости до стадии генерации кода.

---

## 2. Архитектура Компилятора `axlc`

1. **Лексер ([pkg/lexer](file:///home/den/AxiomLang/pkg/lexer/lexer.go))**: Токенизирует исходный код `.axl` в поток токенов.
2. **Парсер Пратта ([pkg/parser](file:///home/den/AxiomLang/pkg/parser/parser.go))**: Строит типизированное AST-дерево ([pkg/ast](file:///home/den/AxiomLang/pkg/ast/ast.go)).
3. **Статический верификатор ([pkg/checker](file:///home/den/AxiomLang/pkg/checker/checker.go))**: Контролирует матрицу мандатов.
4. **LWE Knot Prover ([stdlib/crypto](file:///home/den/AxiomLang/stdlib/crypto/crypto.go))**: Защищает символы бинарного файла криптографической матрицей решеток.
5. **Кодогенератор ([pkg/codegen](file:///home/den/AxiomLang/pkg/codegen/codegen.go))**: Генерирует машинный рантайм.

---

## 3. Стандартная Библиотека

- **[`stdlib/fs`](file:///home/den/AxiomLang/stdlib/fs/fs.go)**: Транзакционная мандатная файловая система с криптографическим хэшированием блоков `Inode`.
- **[`stdlib/actor`](file:///home/den/AxiomLang/stdlib/actor/actor.go)**: Конкурентная модель легковесных акторов без глобальных мьютексов.
- **[`stdlib/gui`](file:///home/den/AxiomLang/stdlib/gui/gui.go)**: Пространственный интерфейс GlassDark 120 FPS.
- **[`stdlib/data`](file:///home/den/AxiomLang/stdlib/data/dataframe.go)**: Zero-copy колоночные структуры данных с прямым обращением к памяти.
- **[`stdlib/crypto`](file:///home/den/AxiomLang/stdlib/crypto/crypto.go)**: Квантово-устойчивые алгоритмы (Ring-LWE).
