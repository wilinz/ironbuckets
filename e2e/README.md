# End-to-End Tests

Playwright-based E2E tests for IronBuckets.

## Setup

```bash
# Install dependencies
npm install

# Install Playwright browsers
npm run playwright:install
```

## Running Tests

```bash
# Run all tests (headless)
npm run test:e2e

# Run with UI mode (interactive)
npm run test:e2e:ui

# Run in headed mode (see browser)
npm run test:e2e:headed

# Debug mode (step through tests)
npm run test:e2e:debug
```

## Environment Variables

Configure the test environment:

```bash
export MINIO_ENDPOINT="localhost:9000"
export ADMIN_USER="minioadmin"
export ADMIN_PASSWORD="minioadmin"
export APP_URL="http://localhost:8080"
```

## Bug Discovery: User Creation Modal

### Issue

The user creation flow has a UX bug where the modal doesn't close after successful user creation.

#### Root Cause

```@/Users/damacus/repos/damacus/ironbuckets/views/partials/user_create_modal.html#10
<form hx-post="/users/create" hx-target="this" hx-swap="outerHTML">
```

The form uses `hx-swap="outerHTML"`, expecting HTML to replace the form element.

```@/Users/damacus/repos/damacus/ironbuckets/internal/handlers/users_handler.go#72
return c.NoContent(http.StatusCreated)
```

However, the handler returns `NoContent`, so HTMX has nothing to swap, leaving the modal open with no feedback.

#### Expected Behavior

1. User submits form
2. Server creates user successfully
3. Modal closes
4. User list refreshes to show new user
5. Success message appears

#### Actual Behavior

1. User submits form
2. Server creates user successfully
3. Modal remains open (no HTML to swap)
4. No visual feedback
5. User must manually close modal or refresh page

#### Proposed Fix

##### Option 1: Return Success HTML Fragment

```go
// Return a success response that triggers modal close and list refresh
return c.HTML(http.StatusCreated, `
  <div hx-get="/users" hx-target="body" hx-swap="innerHTML" hx-trigger="load"></div>
`)
```

##### Option 2: Use HX-Trigger Header

```go
// Trigger a custom event to close modal and refresh list
c.Response().Header().Set("HX-Trigger", "userCreated")
return c.NoContent(http.StatusCreated)
```

Then add JavaScript to listen for the event and refresh the page.

##### Option 3: Use HX-Redirect (Simplest)

```go
// Redirect back to users page (forces full refresh)
c.Response().Header().Set("HX-Redirect", "/users")
return c.NoContent(http.StatusOK)
```

This is the pattern already used in `UploadObject` handler.

## Test Coverage

- ✅ User creation flow (with bug demonstration)
- ✅ Logout functionality
- ✅ Login as newly created user
- ✅ Form validation
- ✅ Modal cancel behavior
- ✅ Modal backdrop click behavior
