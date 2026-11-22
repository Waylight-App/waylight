# Roadmap

Waylight is planned in public. Every item is a GitHub issue; milestones
group them. The order is a guideline, not a promise, and contributions can
move anything forward.

## Shipped

[v0.1.0](https://github.com/Waylight-App/waylight/releases/tag/v0.1.0) is the
first release. Its notes list every change. The milestones
[v1.0](https://github.com/Waylight-App/waylight/milestone/1),
[v1.1](https://github.com/Waylight-App/waylight/milestone/2) and
[v1.2](https://github.com/Waylight-App/waylight/milestone/3) have no open
issues apart from [#90](https://github.com/Waylight-App/waylight/issues/90)
(project health) in v1.0.

## Open

Everything that is open right now, by theme. The
[Backlog milestone](https://github.com/Waylight-App/waylight/milestone/4)
holds ideas without a date. There are no dates for any of this.

**Plain language and help**

- [#699](https://github.com/Waylight-App/waylight/issues/699) Plain words in the interface: jargon pass and word list
- [#678](https://github.com/Waylight-App/waylight/issues/678) Help in the interface on every widget
- [#280](https://github.com/Waylight-App/waylight/issues/280) Tracking: Waylight for non-technical users
- [#618](https://github.com/Waylight-App/waylight/issues/618) Guided setup in place for the remaining integrations

**Places and widgets**

- [#674](https://github.com/Waylight-App/waylight/issues/674) Places everywhere: use my location, stops and waste calendar from a place
- [#695](https://github.com/Waylight-App/waylight/issues/695) Waste: collection calendar by place from a provider

**Performance**

- [#718](https://github.com/Waylight-App/waylight/issues/718) Web entry: load feature services on first use

**Accessibility and translation**

- [#354](https://github.com/Waylight-App/waylight/issues/354) Manual audit with NVDA and VoiceOver
- [#124](https://github.com/Waylight-App/waylight/issues/124) Add a third UI language

**Distribution and hardware**

- [#290](https://github.com/Waylight-App/waylight/issues/290) Browser extension in the Chrome, Edge and Firefox stores
- [#130](https://github.com/Waylight-App/waylight/issues/130) Tracking: Raspberry Pi appliance image
- [#361](https://github.com/Waylight-App/waylight/issues/361) Hardware reports for the Raspberry Pi appliance and kiosk display

**Project**

- [#90](https://github.com/Waylight-App/waylight/issues/90) Visible roadmap, changelog and contributor onboarding

- **What changed:** see the [releases and changelog](https://github.com/Waylight-App/waylight/releases),
  generated from Conventional Commits for every version.
- **Want to help?** Issues labelled
  [good first issue](https://github.com/Waylight-App/waylight/labels/good%20first%20issue)
  list the files and tests to touch. Start with [CONTRIBUTING.md](CONTRIBUTING.md).
- **Have an idea or a question?** Open an issue with the feature form or ask in
  [Discussions](https://github.com/Waylight-App/waylight/discussions).

## Principles that do not change

1. **Private by default:** no telemetry, no outbound request you did not configure.
2. **Least privilege:** read-only access first; anything that can change a system is a separate, optional step ([security model](docs/security-model.md)).
3. **Lean:** one binary, SQLite, small images; new dependencies need a reason.
4. **Clean Architecture:** every feature is a module with the same four rings, enforced in CI.
