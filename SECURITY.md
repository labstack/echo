# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 5.x.x   | :white_check_mark: |
| >= 4.15.x | :white_check_mark: |
| < 4.15   | :x:                |

## Reporting a Vulnerability

The Echo team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Security Vulnerability

**Preferred Method: GitHub Private Vulnerability Reporting**

We recommend using GitHub's Private Vulnerability Reporting feature:

1. Navigate to the [Security tab](https://github.com/labstack/echo/security) of this repository
2. Click "Report a vulnerability" to open a private security advisory
3. Provide detailed information about the vulnerability

This is the preferred method as it allows us to collaborate with you privately to validate and fix the issue before public disclosure.

**Alternative Method: Email**

If you prefer to report via email or if GitHub Private Vulnerability Reporting is not available, please contact the maintainers directly:

- Look for maintainer email addresses in recent verified commits
- Send your report to one or more of the following maintainers:
  - [Martti T. (@aldas)](https://github.com/aldas)
  - [Vishal Rana (@vishr)](https://github.com/vishr)
  - [Roland Lammel (@lammel)](https://github.com/lammel)
  - [Pablo Andres Fuente (@pafuent)](https://github.com/pafuent)

### What to Include in Your Report

To help us triage and fix the issue quickly, please include:

- Description of the vulnerability
- Steps to reproduce the issue
- Affected versions
- Any potential mitigations or workarounds
- Your contact information for follow-up questions

### What to Expect

- We will acknowledge receipt of your vulnerability report within 48 hours
- We will provide a more detailed response within 7 days indicating the next steps
- We will keep you informed about our progress toward a fix
- We will credit you in the security advisory (unless you prefer to remain anonymous)

### Security Update Process

When we receive a security bug report, we will:

1. Confirm the problem and determine affected versions
2. Audit code to find similar problems
3. Prepare fixes for all supported versions
4. Release patches as quickly as possible

Thank you for helping keep Echo and its users safe!
