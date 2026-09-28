import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Checkbox } from "@/components/ui/checkbox"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

export default function StyleGuidePage() {
  return (
    <div className="space-y-8 max-w-5xl mx-auto pb-20">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">Style Guide</h1>
        <p className="text-muted-foreground">
          A showcase of the UI components and design system used in Forge Admin.
        </p>
      </div>

      {/* Colors */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Colors</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-background border shadow-sm" />
            <p className="text-sm font-medium">Background</p>
          </div>
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-card border shadow-sm" />
            <p className="text-sm font-medium">Card</p>
          </div>
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-primary shadow-sm" />
            <p className="text-sm font-medium">Primary</p>
          </div>
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-secondary shadow-sm" />
            <p className="text-sm font-medium">Secondary</p>
          </div>
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-accent shadow-sm" />
            <p className="text-sm font-medium">Accent</p>
          </div>
          <div className="space-y-1.5">
            <div className="h-20 w-full rounded-lg bg-destructive shadow-sm" />
            <p className="text-sm font-medium">Destructive</p>
          </div>
        </div>
      </section>

      {/* Typography */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Typography</h2>
        <Card>
            <CardContent className="p-6 space-y-4">
                <div>
                    <h1 className="text-4xl font-extrabold tracking-tight lg:text-5xl">Heading 1</h1>
                    <p className="text-sm text-muted-foreground">text-4xl font-extrabold tracking-tight lg:text-5xl</p>
                </div>
                <div>
                    <h2 className="text-3xl font-semibold tracking-tight first:mt-0">Heading 2</h2>
                    <p className="text-sm text-muted-foreground">text-3xl font-semibold tracking-tight</p>
                </div>
                <div>
                    <h3 className="text-2xl font-semibold tracking-tight">Heading 3</h3>
                    <p className="text-sm text-muted-foreground">text-2xl font-semibold tracking-tight</p>
                </div>
                <div>
                    <p className="leading-7 [&:not(:first-child)]:mt-6">
                        The king, seeing how much happier his subjects were, realized the error of his ways and repealed the joke tax.
                    </p>
                    <p className="text-sm text-muted-foreground">Regular paragraph (leading-7)</p>
                </div>
            </CardContent>
        </Card>
      </section>

      {/* Buttons */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Buttons</h2>
        <Card>
            <CardContent className="p-6">
                <div className="flex flex-wrap gap-4">
                    <Button>Default</Button>
                    <Button variant="secondary">Secondary</Button>
                    <Button variant="destructive">Destructive</Button>
                    <Button variant="outline">Outline</Button>
                    <Button variant="ghost">Ghost</Button>
                    <Button variant="link">Link</Button>
                </div>
                <div className="flex flex-wrap gap-4 mt-4">
                    <Button size="sm">Small</Button>
                    <Button size="default">Default</Button>
                    <Button size="lg">Large</Button>
                    <Button size="icon">Icon</Button>
                </div>
            </CardContent>
        </Card>
      </section>

      {/* Form Elements */}
      <section className="space-y-4">
        <h2 className="text-xl font-semibold">Form Elements</h2>
        <Card>
            <CardContent className="p-6 space-y-4 max-w-md">
                <div className="space-y-2">
                    <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">Email</label>
                    <Input type="email" placeholder="Email" />
                </div>
                <div className="flex items-center space-x-2">
                    <Checkbox id="terms" />
                    <label
                        htmlFor="terms"
                        className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                        Accept terms and conditions
                    </label>
                </div>
            </CardContent>
        </Card>
      </section>

       {/* Cards */}
       <section className="space-y-4">
        <h2 className="text-xl font-semibold">Cards</h2>
        <div className="grid md:grid-cols-2 gap-4">
            <Card>
                <CardHeader>
                    <CardTitle>Card Title</CardTitle>
                    <CardDescription>Card Description</CardDescription>
                </CardHeader>
                <CardContent>
                    <p>Card Content</p>
                </CardContent>
                <CardFooter>
                    <p>Card Footer</p>
                </CardFooter>
            </Card>
             <Card className="bg-primary text-primary-foreground">
                <CardHeader>
                    <CardTitle>Primary Card</CardTitle>
                    <CardDescription className="text-primary-foreground/80">Card Description</CardDescription>
                </CardHeader>
                <CardContent>
                    <p>Card Content</p>
                </CardContent>
            </Card>
        </div>
      </section>
      
      {/* Table */}
       <section className="space-y-4">
        <h2 className="text-xl font-semibold">Table</h2>
        <Card>
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-[100px]">Invoice</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Method</TableHead>
                        <TableHead className="text-right">Amount</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    <TableRow>
                        <TableCell className="font-medium">INV001</TableCell>
                        <TableCell>Paid</TableCell>
                        <TableCell>Credit Card</TableCell>
                        <TableCell className="text-right">$250.00</TableCell>
                    </TableRow>
                    <TableRow>
                        <TableCell className="font-medium">INV002</TableCell>
                        <TableCell>Pending</TableCell>
                        <TableCell>PayPal</TableCell>
                        <TableCell className="text-right">$150.00</TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </Card>
      </section>

    </div>
  )
}
