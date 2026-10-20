package gui_test

import (
	"testing"

	"AxiomLang/stdlib/gui"
)

func TestAXLUIModel_ImportAndGenerate(t *testing.T) {
	model := gui.NewAXLUIModel("test_window.axlui", "Test Window")

	// 1. Verify default model code generation
	code := model.GenerateAxiomCode()
	if len(code) == 0 {
		t.Fatalf("expected non-empty generated axiom code")
	}

	// 2. Test HTML/CSS Import
	rawHTML := `
<div>
  <h1>Заголовок дашборда</h1>
  <button class="primary">Запустить расчет</button>
</div>
`
	err := model.ImportHTMLCSS(rawHTML)
	if err != nil {
		t.Fatalf("failed to import HTML/CSS: %v", err)
	}

	if len(model.Root.Children) != 2 {
		t.Fatalf("expected 2 imported children, got %d", len(model.Root.Children))
	}

	firstText := model.Root.Children[0].Text
	if firstText != "Заголовок дашборда" {
		t.Fatalf("expected text 'Заголовок дашборда', got %s", firstText)
	}

	btnText := model.Root.Children[1].Text
	if btnText != "Запустить расчет" {
		t.Fatalf("expected button text 'Запустить расчет', got %s", btnText)
	}
}
