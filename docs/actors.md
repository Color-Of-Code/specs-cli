# Actors

The toolchain is built around a short, linear authoring chain. Four actors
cover everything the engine is concerned with; one person may hold several
of these roles in the same project.

Setup, review, and framework-distribution work happens **outside** this
chain and is described as *operational roles* in [roles.md](roles.md).

```text
Stakeholder ──► Author ──► Analyst ──► Architect
   input        requirements   features    components / services / APIs
```

Each actor refines the work of the previous one. Their output lives in
canonical locations under `model/`, drafted inside change requests.

## Stakeholder

Provides input. The stakeholder describes a need, a problem, or a desired
behaviour by **opening a change request**. They do not draft requirements
themselves.

- Opens a change request: `specs cr new --id <NNN> --slug <slug>`.
- Writes informal notes, examples, and acceptance criteria inside the CR
  folder.

## Author

Turns stakeholder input into properly formed **requirements**. The author
owns wording, IDs, and scope of each requirement.

- Scaffolds a requirement: `specs scaffold requirement --cr <NNN> <path>`.
- Edits the markdown to fit the template structure.
- Runs `specs format` and `specs lint` while iterating.

## Analyst

Derives **features** from approved requirements. The analyst groups related
requirements into features and ensures every requirement is implemented by
at least one feature.

- Scaffolds a feature: `specs scaffold feature --cr <NNN> <path>`.
- Lists implementing features under each requirement and the requirements
  under each feature.
- Runs `specs link check` to confirm the bidirectional links are symmetric.

## Architect

Decomposes features into **components, services, and APIs**. The architect
also keeps component baselines aligned with their upstream repositories.

- Scaffolds artifacts:
  `specs scaffold component|service|api --cr <NNN> <path>`.
- Maintains the components table: `specs baseline update`.
- Verifies traceability: `specs link check`,
  `specs visualize traceability`.

## How they work together

1. **Stakeholder** opens a CR with raw input.
2. **Author** drafts requirements inside the CR.
3. **Analyst** adds features that implement those requirements.
4. **Architect** adds components, services, and APIs implementing the
   features.
5. Anyone runs `specs format`, `specs lint`, and `specs link check` to
   validate the CR.
6. The CR is drained (`specs cr drain --id <NNN>`) and merged: files move
   into their canonical homes under `model/`.

Setup tasks (installing the engine, initialising a host repo, managing
the framework registry) are one-time prerequisites and are not part of
the authoring chain. See [overview.md](overview.md) and
[install.md](install.md).
