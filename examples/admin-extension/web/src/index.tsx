// Entry point for Forge Admin extension.
// The default export function is called with the host admin's ForgeAdminAPI instance.
export default function (api: any) {
  const React = api.react;
  const { Card, CardHeader, CardTitle, CardContent, Button } = api.ui;

  function AnalyticsWidget() {
    const [count, setCount] = React.useState(0);

    return (
      <Card>
        <CardHeader>
          <CardTitle>Analytics Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <p>Active sessions: {count}</p>
          <Button onClick={() => setCount((c: number) => c + 1)}>
            Increment
          </Button>
        </CardContent>
      </Card>
    );
  }

  function AnalyticsPage() {
    return (
      <div style={{ padding: "1.5rem" }}>
        <h1>Analytics Dashboard</h1>
        <p>Custom analytics extension page loaded inside the admin shell.</p>
      </div>
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
