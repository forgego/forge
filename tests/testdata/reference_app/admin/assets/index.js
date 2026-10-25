export default function (api) {
  function ReferenceOrdersSummary(props) {
    return api.react.createElement(
      api.ui.Card,
      { "data-testid": "ref-widget" },
      api.react.createElement(
        api.ui.CardHeader,
        null,
        api.react.createElement(api.ui.CardTitle, null, "Orders Summary")
      ),
      api.react.createElement(
        api.ui.CardContent,
        null,
        api.react.createElement(
          api.ui.CardDescription,
          null,
          "Reference orders summary widget"
        )
      )
    );
  }

  function ApprovalsPage(props) {
    return api.react.createElement(
      api.ui.Card,
      { "data-testid": "ref-approvals-page" },
      api.react.createElement(api.ui.PageHeader, {
        title: "Approvals",
        description: "Review pending orders and customer approvals",
      }),
      api.react.createElement(
        api.ui.CardContent,
        null,
        api.react.createElement(
          api.ui.CardDescription,
          null,
          "Reference approvals page content"
        )
      )
    );
  }

  function CustomerFieldEditor(props) {
    var value =
      props && props.value != null && props.value !== ""
        ? String(props.value)
        : undefined;

    return api.react.createElement(
      api.ui.Select,
      {
        value: value,
        onValueChange: function (nextVal) {
          if (props && typeof props.onChange === "function") {
            props.onChange(nextVal);
          }
        },
      },
      api.react.createElement(
        api.ui.SelectTrigger,
        { "data-testid": "ref-customer-editor" },
        api.react.createElement(api.ui.SelectValue, {
          placeholder: "Select customer",
        })
      ),
      api.react.createElement(
        api.ui.SelectContent,
        null,
        api.react.createElement(api.ui.SelectItem, { value: "1" }, "Customer 1"),
        api.react.createElement(api.ui.SelectItem, { value: "2" }, "Customer 2")
      )
    );
  }

  api.registerWidget("reference-orders-summary", ReferenceOrdersSummary);
  api.registerPage(
    { path: "approvals", title: "Approvals", nav: true },
    ApprovalsPage
  );
  api.registerFieldEditor(
    { model: "orders", field: "customer" },
    CustomerFieldEditor
  );
}
