## Description
A concise summary of the changes proposed in this pull request.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Documentation update
- [ ] Refactoring / Code cleanup
- [ ] CI/CD or Release infrastructure

## Related Issues
Fixes #(issue number)

## Architectural & Safety Checklist
- [ ] Tested on Windows 64-bit (`amd64`).
- [ ] Tested on Windows 32-bit (`386`) compatibility.
- [ ] Fallback to System Proxy remains intact if Wintun fails.
- [ ] Disconnect cleans up routes, DNS cache, and system proxy settings.
- [ ] All automated tests pass (`go test ./...`).
- [ ] No hardcoded secrets, personal tokens, or passwords introduced.

## Screenshots / Verification Output (Optional)
Add any relevant logs or screenshots demonstrating the change.
