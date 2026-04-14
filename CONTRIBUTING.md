# Contributing to CallSign PBX

Thank you for your interest in contributing to CallSign PBX! This document provides guidelines and best practices for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourname/callsign.git`
3. Add upstream: `git remote add upstream https://github.com/original/callsign.git`
4. Create a branch: `git checkout -b feature/your-feature`

## Development Workflow

### 1. Setup Environment

```bash
# Copy environment files
cp .env.example .env
cp api/.env.example api/.env

# Start infrastructure
docker-compose up -d postgres redis

# Install dependencies
cd api && go mod download
cd ../ui && npm install
```

### 2. Make Changes

- Write clear, focused commits
- Follow existing code style
- Add tests for new features
- Update documentation

### 3. Test Your Changes

```bash
# Backend tests
cd api
go test ./...
golangci-lint run

# Frontend tests
cd ui
npm run lint
npm test

# Build verification
npm run build
```

### 4. Commit Guidelines

Use conventional commits:

```
feat: add new widget management feature
fix: resolve memory leak in queue processor
docs: update API reference
refactor: simplify dialplan generation
test: add tests for extension handlers
chore: update dependencies
```

### 5. Submit Pull Request

1. Push to your fork: `git push origin feature/your-feature`
2. Create PR against `main` branch
3. Fill out PR template
4. Link related issues: `Fixes #123`
5. Request review from maintainers

## Coding Standards

### Go (Backend)

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `golint` for linting
- Document exported functions
- Handle all errors explicitly

```go
// Good
func (h *Handler) GetWidget(c *fiber.Ctx) error {
    id := c.Params("id")
    widget, err := h.getWidgetByID(id)
    if err != nil {
        return utils.Error(c, fiber.StatusNotFound, "Widget not found")
    }
    return utils.Success(c, widget)
}

// Bad
func (h *Handler) GetWidget(c *fiber.Ctx) error {
    widget, _ := h.getWidgetByID(c.Params("id"))
    return c.JSON(widget)
}
```

### Vue (Frontend)

- Use Composition API
- Follow component naming conventions
- Use CSS variables from `variables.css`
- Add JSDoc comments for composables

```vue
<!-- Component naming -->
<!-- Good -->
<WidgetManager />
<BaseButton />

<!-- Bad -->
<widgetManager />
<Button />
```

### CSS

- Use CSS variables for colors/sizing
- Follow BEM methodology
- Mobile-first responsive design

```css
/* Good */
.widget-card {
  background: var(--bg-card);
  padding: var(--spacing-4);
}

.widget-card__title {
  font-size: var(--text-lg);
}

/* Bad */
.widgetCard {
  background: #fff;
  padding: 16px;
}
```

## Testing Requirements

### Backend

- Unit tests for handlers
- Integration tests for API endpoints
- Table-driven tests preferred

```go
func TestCreateWidget(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateWidgetRequest
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Frontend

- Unit tests for components
- Tests for composables
- E2E tests for critical paths

```javascript
import { mount } from '@vue/test-utils'
import WidgetForm from './WidgetForm.vue'

describe('WidgetForm', () => {
  it('emits submit with form data', async () => {
    const wrapper = mount(WidgetForm)
    await wrapper.find('input').setValue('Test')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toBeTruthy()
  })
})
```

## Documentation

Update documentation when:
- Adding new API endpoints
- Changing architecture
- Adding new features
- Modifying setup process

### Documentation Structure

```
docs/
├── API_REFERENCE.md      # API docs (auto-generated from comments)
├── ARCHITECTURE.md       # System design
├── FRONTEND.md          # Frontend patterns
└── SETUP_AND_USAGE.md   # Installation guide
```

### Skills

For repeated patterns, add to `.claude/skills/`:

```markdown
# Skill Name

## Overview
Description of the skill.

## Common Patterns
Examples and explanations.

## Testing
How to verify.

## Resources
Links to docs.
```

## Review Process

1. Automated checks (lint, test, build)
2. Code review by maintainers
3. Manual testing for UI changes
4. Documentation review

### Review Checklist

- [ ] Code follows style guide
- [ ] Tests pass
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
- [ ] Security considerations addressed

## Release Process

1. Create release branch: `release/v1.2.3`
2. Update CHANGELOG.md
3. Update version numbers
4. Create PR to main
5. Tag release after merge: `git tag v1.2.3`
6. Deploy to production

## Questions?

- Check existing [documentation](docs/)
- Search [GitHub Issues](https://github.com/yourorg/callsign/issues)
- Ask in [Discussions](https://github.com/yourorg/callsign/discussions)

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Added to project credits

Thank you for contributing to CallSign PBX!
