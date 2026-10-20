package gui_test

import (
	"testing"

	"AxiomLang/stdlib/gui"
)

func TestHighLevelGUI_FluentBuilder(t *testing.T) {
	// Создание сложнейшего космического дашборда в 3 строки:
	app := gui.Window("Quantum AI Cluster Control",
		gui.Header("🛡️ Quantum Security Dashboard", "SYSTEM NOMINAL"),
		gui.Row(
			gui.MetricBadge("CPU Core 0", "98.4%", "+2.1%"),
			gui.MetricBadge("Memory Bandwidth", "1.2 TB/s", "MAX"),
			gui.MetricBadge("LWE Lattice Security", "2048 Bit", "SEALED"),
			gui.MetricBadge("Actor Mailbox Throughput", "28.4M/s", "+14%"),
		),
		gui.Grid(
			gui.Card("Neural Invariant Guard", "Все инварианты верифицированы без сбоев."),
			gui.Card("Zero-Ambient Controller", "Мандаты изолированы на аппаратном уровне."),
			gui.Card("Bare-Metal Kernel Telemetry", "Ring 0 время отклика: 0.12 мкс."),
		),
		gui.Row(
			gui.Button("Перезапустить Акторы", "rose", "restart_actors"),
			gui.Button("Сгенерировать LWE Ключ", "emerald", "gen_key"),
			gui.Button("Экспорт в LLVM IR", "cyan", "export_llvm"),
		),
	)

	if len(app.Root.Children) != 4 {
		t.Fatalf("expected 4 root containers, got %d", len(app.Root.Children))
	}

	header := app.Root.Children[0]
	if len(header.Children) != 2 {
		t.Fatalf("expected header to contain title and badge")
	}

	code := app.GenerateAxiomCode()
	if len(code) == 0 {
		t.Fatalf("expected generated code from high-level builder")
	}
}
