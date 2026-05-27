# Message Chains — Hide Delegate (Law of Demeter)

## What
Long chains of method calls navigating through multiple objects: `a.b().c().d()`. Violates the Law of Demeter — "only talk to your immediate friends."

## Why It's Harmful
- **Brittle coupling**: A change to any intermediate object breaks the entire chain. Each dot is a dependency.
- **AI confusion**: LLMs struggle to track types through 3+ levels of indirection. Bugs appear when models assume wrong intermediate types.
- **Test complexity**: Mocking a 5-level chain requires setting up stubs at every level. Tests become fragile and verbose.

## Step-by-Step Refactoring

### 1. Identify the Chain and the Real Target
```go
// BEFORE: Message chain
func getManagerName(company *Company, deptID string) string {
    return company.GetDepartment(deptID).GetManager().GetName()
}

func getManagerOffice(company *Company, deptID string) string {
    return company.GetDepartment(deptID).GetManager().GetOffice().GetAddress()
}
```

### 2. Hide the Delegate — Create a Wrapper Method
```go
// AFTER: Hide the delegate
func (c *Company) GetManagerName(deptID string) string {
    return c.GetDepartment(deptID).GetManager().GetName()
}

// Client code becomes:
name := company.GetManagerName("sales")
```

### 3. For Deeper Chains — Introduce a Dedicated Query Object
```go
// BEFORE: Deep chain
func getOfficeAddress(company *Company, deptID string) string {
    return company.
        GetDepartment(deptID).
        GetManager().
        GetOffice().
        GetAddress().
        Format()
}

// AFTER: Query object
type ManagerInfo struct {
    Name        string
    OfficeAddr  string
    DirectCount int
}

func (c *Company) GetManagerInfo(deptID string) (ManagerInfo, error) {
    dept, err := c.GetDepartment(deptID)
    if err != nil {
        return ManagerInfo{}, err
    }
    mgr := dept.GetManager()
    return ManagerInfo{
        Name:       mgr.GetName(),
        OfficeAddr: mgr.GetOffice().GetAddress(),
        DirectCount: len(mgr.GetDirectReports()),
    }, nil
}
```

### 4. Validate Nil Checks Along the Chain
Each intermediate call is a potential nil — the query object approach forces explicit error handling at boundaries.

## Verification
- [ ] No chain exceeds 2 dots (a.b() is fine, a.b().c() needs justification)
- [ ] Intermediate structures do not leak to callers
- [ ] Nil/null checks are explicit and at service boundaries
- [ ] Re-run `analyze_code` — Message Chains smell should be resolved
