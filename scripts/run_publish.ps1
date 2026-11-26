$body = Get-Content -Raw -Encoding UTF8 "C:\Users\Yousef Kia\Desktop\FYRNOX\scripts\bilingual_issue_report.md"
& "C:\Users\Yousef Kia\Desktop\FYRNOX\scripts\publish_github_issue.ps1" -Title "[Feature] Implement Firefox-Compatible TLS Control Plane (uTLS), Unified Transport Factory & 4-Cell A/B Matrix Diagnostic" -BodyMarkdown $body
