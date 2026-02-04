import { useState } from "react";
import { Clock, FormInput, Palette } from "lucide-react";

import AdminLayout from "../components/layout/AdminLayout";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Checkbox } from "../components/ui/checkbox";
import { Input } from "../components/ui/input";
import { Switch } from "../components/ui/switch";

export default function FormPlaygroundPage() {
  const [switchOn, setSwitchOn] = useState(true);
  const [status, setStatus] = useState("draft");
  const [color, setColor] = useState("#6366f1");
  const [timeValue, setTimeValue] = useState("09:30");

  return (
    <AdminLayout>
      <div className="space-y-6">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">Form Playground</h1>
          <p className="text-sm text-muted-foreground">
            A reference page for switches, time fields, and custom inputs used in
            create/edit experiences.
          </p>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <Card className="glass-lite border-border/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <FormInput className="h-4 w-4 text-primary" />
                Primary inputs
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Title
                </label>
                <Input placeholder="Campaign name" />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Email
                </label>
                <Input type="email" placeholder="team@forge.dev" />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Status
                </label>
                <select
                  className="flex h-10 w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20"
                  value={status}
                  onChange={(e) => setStatus(e.target.value)}
                >
                  <option value="draft">Draft</option>
                  <option value="active">Active</option>
                  <option value="paused">Paused</option>
                </select>
              </div>
              <div className="flex items-center gap-3 rounded-lg border border-border/50 bg-muted/20 p-3">
                <Checkbox checked />
                <span className="text-sm font-medium">Enable notifications</span>
              </div>
            </CardContent>
          </Card>

          <Card className="glass-lite border-border/50">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Clock className="h-4 w-4 text-primary" />
                Toggles & time
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-lg border border-border/50 bg-muted/20 p-3">
                <div>
                  <p className="text-sm font-semibold">Auto-publish</p>
                  <p className="text-xs text-muted-foreground">
                    Launch automatically when schedule is met.
                  </p>
                </div>
                <Switch checked={switchOn} onCheckedChange={setSwitchOn} />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Schedule time
                </label>
                <Input
                  type="time"
                  value={timeValue}
                  onChange={(e) => setTimeValue(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Date & time
                </label>
                <Input type="datetime-local" />
              </div>
            </CardContent>
          </Card>

          <Card className="glass-lite border-border/50 lg:col-span-2">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-lg">
                <Palette className="h-4 w-4 text-primary" />
                Custom fields
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Accent color
                </label>
                <div className="flex items-center gap-3">
                  <Input
                    type="color"
                    value={color}
                    onChange={(e) => setColor(e.target.value)}
                    className="h-10 w-16 p-1"
                  />
                  <Input value={color} readOnly />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Budget
                </label>
                <Input type="number" placeholder="5000" />
              </div>
              <div className="space-y-2 md:col-span-2">
                <label className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                  Notes
                </label>
                <textarea
                  className="flex min-h-[120px] w-full rounded-lg border border-border/50 bg-background/50 px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20"
                  placeholder="Add context for your team..."
                />
              </div>
              <div className="flex flex-wrap gap-2 md:col-span-2">
                <Button>Primary</Button>
                <Button variant="secondary">Secondary</Button>
                <Button variant="outline">Outline</Button>
                <Button variant="ghost">Ghost</Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </AdminLayout>
  );
}
