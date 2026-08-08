# Bundled Agent Skills

These skill bundles are embedded into the `osmedeus` binary and installed by
`osmedeus skills install`. Because they ship with the binary, an installed
skill always matches the version of osmedeus that wrote it.

## This directory is the source of truth

Skills are **authored here**, alongside the code they document, so a change to a
step type or CLI flag lands in the same commit as the skill text describing it.

The standalone repo is a published mirror:

    https://github.com/osmedeus/osmedeus-skills

It exists so agents can install a skill without osmedeus (`bunx skills add ...`).
Push this directory out to a local checkout of it with:

```bash
# defaults to ../osmedeus-skills
make sync-skills

# or point at the checkout explicitly
make sync-skills SKILLS_DEST=/path/to/osmedeus-skills
```

The sync copies every top-level directory containing a `SKILL.md`, so adding a
new bundle here requires no code change. It only writes bundle directories —
review and commit in the skills repo yourself. Its own `README.md` is the public
landing page and is deliberately left alone.

## Layout

Each bundle is a directory whose name is its canonical identity — it is what
`skills install` names the destination folder and what CLI arguments refer to:

```
skills/
└── osmedeus-expert/
    ├── SKILL.md          # frontmatter (name, description) + main content
    └── references/       # progressive-disclosure detail, loaded on demand
        └── *.md
```
