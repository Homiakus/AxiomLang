# Стратегия и Архитектура создания Нейросетей на AxiomLang (`axl`)

**Цель:** Обеспечить компиляцию, скульптурирование (Model Sculptor) и исполнение нейросетевых архитектур непосредственно на языке **AxiomLang** с сохранением математических гарантий и субмикросекундной задержки.

---

## 1. Фундаментальный принцип: «CPU Governs; Accelerators Compute»

В современных фреймворках (PyTorch, TensorFlow, TensorRT/CUDA) рантайм нейросети требует неограниченного доступа к системной памяти, драйверам GPU и огромного стека C++/Python библиотек.

В **AxiomLang** действует модель типизированного владения:
```text
         AxiomLang Intent / Plan
                   │
                   ▼
  CPU Capability Authority (cap[Accelerator.Queue])
                   │
                   ▼
  Zero-Copy Tensor Graph (L1 CPU Cache / Bounded VRAM)
                   │
                   ▼
  Evidence & Proof Seal Verification (Fail-Closed)
```

1. **Мандатное исполнение (Capability-Bound Compute):**
   * Нейросеть не может самостоятельно выделить память на GPU или прочитать чужие тензоры без явного мандата `cap[Tensor.Memory]`.
2. **Нулевой оверхед сериализации (Zero-Copy Interop):**
   * Тензоры в `stdlib/nn` хранятся в непрерывных выровненных буферах, совместимых напрямую с SIMD CPU (AVX-512/NEON) и DMA очередями ускорителей.
3. **Модели, скомпилированные в L1 кэш (Spectrum Swarm Paradigm):**
   * Архитектуры узких специалистов (до 100M параметров / квантованные микромодели) компилируются в нативный код AxiomLang, помещаясь целиком в L1 кэш инструкций и данных (задержка инференса < 250 нс).

---

## 2. Как выглядит создание и инференс нейросети на AxiomLang

### Пример: Описание и компиляция модели
```axiom
// Запрашиваем мандаты на аппаратное ускорение и чтение весов
require cap[Accelerator.Compute]
require cap[FileSystem.Read("/models/spectrum_expert.axm")]

intent "Сверхбыстрый инференс когнитивного эксперта" {
    // Создаем архитектуру сети через stdlib/nn
    let model = nn.ModelArchitect("SpectrumExpert")
    model.add(nn.Linear(128, 64))
    model.add(nn.Activation.GELU)
    model.add(nn.Linear(64, 10))

    // Векторный тензор входных признаков (Zero-copy)
    let input = nn.Tensor([128], sensor_data)
    
    // Предсказание с субмикросекундной задержкой
    let result = model.predict(input)
    print("Результат инференса:", result)
}
```

---

## 3. Интеграция с роем Spectrum и Tao-Science Sculptor

Нейросети, синтезируемые через платформу скульптурирования (`tao-science` / `Model Sculptor`):
1. **ModelIR Transpiler:** Автоматически трансформирует вычислительные графы ONNX / ModelIR в оптимизированный байткод AxiomLang.
2. **Homomorphic Model Protection:** Веса и топология слоев шифруются на криптографических решетках LWE. Веса невозможно извлечь из скомпилированного бинарника без ключа среды выполнения.
3. **In-Kernel Direct Execution:** Микромодели могут исполняться внутри системных демонов (`rootd`, `axiomd`, планировщиков) без создания внешних процессов.
