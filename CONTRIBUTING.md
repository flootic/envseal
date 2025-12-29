# Contributing to EnvSeal

First off, thanks for taking the time to contribute! ❤️

All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways to help and details about how this project handles them. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The community looks forward to your contributions. 🎉

> And if you like the project, but just don't have time to contribute, that's fine. There are other easy ways to support the project and show your appreciation, which we would also be very happy about:
>
> - Star the project
> - Tweet about it
> - Refer this project in your project's readme
> - Mention the project at local meetups and tell your friends/colleagues

## Table of Contents

- [Contributing to EnvSeal](#contributing-to-envseal)
  - [Table of Contents](#table-of-contents)
  - [Code of Conduct](#code-of-conduct)
  - [I Want To Contribute](#i-want-to-contribute)
    - [Reporting Bugs](#reporting-bugs)
    - [Suggesting Enhancements](#suggesting-enhancements)
      - [How Do I Submit a Good Enhancement Suggestion?](#how-do-i-submit-a-good-enhancement-suggestion)
      - [Before Submitting an Enhancement](#before-submitting-an-enhancement)
  - [I Have a Question](#i-have-a-question)
  - [Styleguides](#styleguides)
    - [Commit Messages](#commit-messages)
    - [Branches](#branches)
    - [Go Style](#go-style)
  - [Testing](#testing)
  - [Project Structure](#project-structure)
  - [Getting Help](#getting-help)
  - [License](#license)

## Code of Conduct

This project and everyone participating in it is governed by the
[EnvSeal Code of Conduct](https://github.com/flootic/envseal/blob//CODE_OF_CONDUCT.md).

## I Want To Contribute

> When contributing to this project, you must agree that you have authored 100% of the content, that you have the necessary rights to the content and that the content you contribute may be provided under the project licence.

### Reporting Bugs

A good bug report shouldn't leave others needing to chase you up for more information. Therefore, we ask you to investigate carefully, collect information and describe the issue in detail in your report. Please complete the following steps in advance to help us fix any potential bug as fast as possible.

- Make sure that you are using the latest version.
- Determine if your bug is really a bug and not an error on your side e.g. using incompatible environment components/versions (Make sure that you have read the [documentation](https://github.com/flootic/envseal/tree/main/docs). If you are looking for support, you might want to check [this section](#i-have-a-question)).
- To see if other users have experienced (and potentially already solved) the same issue you are having, check if there is not already a bug report existing for your bug or error in the [bug tracker](https://github.com/flootic/envseal/issues?q=label%3Abug).
- Also make sure to search the internet (including Stack Overflow) to see if users outside of the GitHub community have discussed the issue.
- Collect information about the bug:
- Stack trace (Traceback)
- OS, Platform and Version (Windows, Linux, macOS, x86, ARM)
- Version of the interpreter, compiler, SDK, runtime environment, package manager, depending on what seems relevant.
- Possibly your input and the output
- Can you reliably reproduce the issue? And can you also reproduce it with older versions?

We use GitHub issues to track bugs and errors. If you run into an issue with the project:

- Open an [Issue](https://github.com/flootic/envseal/issues/new). (Since we can't be sure at this point whether it is a bug or not, we ask you not to talk about a bug yet and not to label the issue.)
- Explain the behavior you would expect and the actual behavior.
- Please provide as much context as possible and describe the *reproduction steps* that someone else can follow to recreate the issue on their own. This usually includes your code. For good bug reports you should isolate the problem and create a reduced test case.
- Provide the information you collected in the previous section.

Once it's filed:

- The project team will label the issue accordingly.
- A team member will try to reproduce the issue with your provided steps. If there are no reproduction steps or no obvious way to reproduce the issue, the team will ask you for those steps and mark the issue as `needs-repro`. Bugs with the `needs-repro` tag will not be addressed until they are reproduced.
- If the team is able to reproduce the issue, it will be marked `needs-fix`, as well as possibly other tags (such as `critical`, `good-first-issue`, etc.) depending on the nature of the issue.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for EnvSeal, **including completely new features and minor improvements to existing functionality**. Following these guidelines will help maintainers and the community to understand your suggestion and find related suggestions.

#### How Do I Submit a Good Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://github.com/flootic/envseal/issues).

- Use a **clear and descriptive title** for the issue to identify the suggestion.
- Provide a **step-by-step description of the suggested enhancement** in as many details as possible.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why. At this point you can also tell which alternatives do not work for you.
- You may want to **include screenshots or screen recordings** which help you demonstrate the steps or point out the part which the suggestion is related to. You can use [LICEcap](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and the built-in [screen recorder in GNOME](https://help.gnome.org/users/gnome-help/stable/screen-shot-record.html.en) or [SimpleScreenRecorder](https://github.com/MaartenBaert/ssr) on Linux.

#### Before Submitting an Enhancement

- Make sure that you are using the latest version.
- Read the [documentation](https://github.com/flootic/envseal/tree/main/docs) carefully and find out if the functionality is already covered, maybe by an individual configuration.
- Perform a [search](https://github.com/flootic/envseal/issues) to see if the enhancement has already been suggested. If it has, add a comment to the existing issue instead of opening a new one.
- Find out whether your idea fits with the scope and aims of the project. It's up to you to make a strong case to convince the project's developers of the merits of this feature. Keep in mind that we want features that will be useful to the majority of our users and not just a small subset. If you're just targeting a minority of users, consider writing an add-on/plugin library.

## I Have a Question

> If you want to ask a question, we assume that you have read the available [Documentation](https://github.com/flootic/envseal/tree/main/docs).

Before you ask a question, it is best to search for existing [Issues](https://github.com/flootic/envseal/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue. It is also advisable to search the internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [Issue](https://github.com/flootic/envseal/issues/new).
- Provide as much context as you can about what you're running into.
- Provide project and platform versions (nodejs, npm, etc), depending on what seems relevant.

We will then take care of the issue as soon as possible.

<!--
You might want to create a separate issue tag for questions and include it in this description. People should then tag their issues accordingly.

Depending on how large the project is, you may want to outsource the questioning, e.g. to Stack Overflow or Gitter. You may add additional contact and information possibilities:
- IRC
- Slack
- Gitter
- Stack Overflow tag
- Blog
- FAQ
- Roadmap
- E-Mail List
- Forum
-->

## Styleguides

### Commit Messages

- Follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification
- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the subject line to 72 characters
- Reference issues and pull requests liberally after the subject line

### Branches

- Use feature branches for new features and bug fixes
- Name branches descriptively (e.g., `feature/enhance-encryption`, `bugfix/config-load-error`)
- Keep branches up to date with the main branch

### Go Style

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines.
- Use `gofmt` to format your code before committing.
- Write clear and concise comments for exported functions, types, and packages.
- Use meaningful variable and function names.
- Avoid global variables where possible.
- Handle errors explicitly and return them to the caller.
- Use interfaces to define behavior and promote decoupling.
- Keep functions small and focused on a single task.

## Testing

- Write tests for new functionality
- Ensure all tests pass before submitting a PR
- Aim for reasonable code coverage
- Tests should be in the same package with `_test.go` suffix

## Project Structure

- `cmd/cli/` - Command-line interface entry point
- `internal/commands/` - CLI command implementations
- `internal/config/` - Configuration and manifest handling
- `internal/crypto/` - Cryptographic operations
- `pkg/` - Reusable packages and libraries

## Getting Help

- Check existing issues and discussions for answers
- Review the `README.md` for usage information
- Open a discussion for questions or general feedback
- Feel free to ask in pull request comments

## License

By contributing to EnvSeal, you agree that your contributions will be licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

Thank you for contributing to EnvSeal! 🙏
