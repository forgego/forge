export default function (api) {
  const React = api.react;
  const { Card, CardHeader, CardTitle, CardContent, Button } = api.ui;

  function AnalyticsWidget() {
    const [count, setCount] = React.useState(0);

    return React.createElement(
      Card,
      null,
      React.createElement(
        CardHeader,
        null,
        React.createElement(CardTitle, null, "Analytics Summary")
      ),
      React.createElement(
        CardContent,
        null,
        React.createElement("p", null, "Active sessions: ", count),
        React.createElement(
          Button,
          { onClick: () => setCount((c) => c + 1) },
          "Increment"
        )
      )
    );
  }

  function AnalyticsPage() {
    return React.createElement(
      "div",
      { style: { padding: "1.5rem" } },
      React.createElement("h1", null, "Analytics Dashboard"),
      React.createElement(
        "p",
        null,
        "Custom analytics extension page loaded inside the admin shell."
      )
    );
  }

  api.registerWidget("analytics-summary", AnalyticsWidget);
  api.registerPage(
    {
      path: "analytics",
      title: "Analytics",
      nav: { label: "Analytics", order: 10 },
    },
    AnalyticsPage
  );
}
