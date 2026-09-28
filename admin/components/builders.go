package components

// Page creates a new page component
func Page(title string) Component {
	return NewComponent(TypePage).
		WithProp("title", title)
}

// Card creates a new card component
func Card(title string) Component {
	return NewComponent(TypeCard).
		WithProp("title", title)
}

// Grid creates a new grid layout
func Grid(cols int) Component {
	return NewComponent(TypeGrid).
		WithProp("columns", cols).
		WithProp("gap", 4)
}

// Text creates a text component
func Text(content string) Component {
	return NewComponent(TypeText).
		WithProp("content", content)
}

// Button creates a button component
func Button(label string, action string) Component {
	return NewComponent(TypeButton).
		WithProp("label", label).
		WithProp("action", action)
}

// Stats creates a stats widget
func Stats(label string, value string, change string) Component {
	return NewComponent(TypeStats).
		WithProp("label", label).
		WithProp("value", value).
		WithProp("change", change)
}

// Chart creates a chart component
func Chart(title string, chartType string, data interface{}) Component {
	return NewComponent(TypeChart).
		WithProp("title", title).
		WithProp("chartType", chartType).
		WithProp("data", data)
}
