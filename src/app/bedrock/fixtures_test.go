package bedrock_test

// fullYAML is a realistic complete config used as the baseline fixture.
const fullYAML = `
actions:
  komp-2:
    cmd: "ffmpeg -i {{.path}} -q:v 2 {{.dir}}/{{.stem}}.mp4"
    when: "isVideo && isLarge && !isHidden"
  komp-18:
    cmd: "ffmpeg -i {{.path}} -q:v 18 {{.dir}}/{{.stem}}.mp4"
    when: "isVideo && size > 50MB"
  upload-s3:
    cmd: "aws s3 cp {{.dir}}/{{.stem}}.mp4 s3://mybucket/"
    when: "exists({{.dir}}/{{.stem}}.mp4)"
  thumbnail:
    cmd: "ffmpeg -i {{.path}} -vf 'thumbnail' -frames:v 1 {{.dir}}/{{.stem}}.jpg"
    when: "isVideo"

pipelines:
  video-workflow:
    steps:
      - komp-2
      - thumbnail
      - upload-s3
  quick-transcode:
    steps:
      - komp-18
      - upload-s3

flags:
  short:
    overrides:
      cmds:
        walk:
          foo: F
        sprint:
          bar: Z
  invoke:
    cmds:
      any:
        files: 2
        folders: 1
  component:
    sampler:
      files: 2
      folders: 1

interaction:
  tui:
    per-item-delay: "1s"

advanced:
  abort-on-error: false
  overwrite-on-collision: false
  extensions:
    suffixes-csv: "jpg,jpeg,png"
    transforms-csv: lower
    map:
      jpeg: jpg

logging:
  max-size-in-mb: 10
  max-backups: 3
  max-age-in-days: 30
  level: info
  time-format: "2006-01-02 15:04:05"
`

// minimalYAML has only the required sections at their zero/default values.
const minimalYAML = `
logging:
  level: info
`

// badLogLevelYAML triggers a validation failure.
const badLogLevelYAML = `
logging:
  level: verbose
`

// negativeDurationYAML triggers interaction validation.
const negativeDurationYAML = `
interaction:
  tui:
    per-item-delay: "-1s"
logging:
  level: info
`

// badShortYAML triggers flag-short validation.
const badShortYAML = `
flags:
  short:
    overrides:
      cmds:
        walk:
          foo: "FF"
logging:
  level: info
`

// missingActionYAML has a pipeline step referencing a nonexistent action.
const missingActionYAML = `
actions:
  komp-2:
    cmd: "ffmpeg -i {{.path}} -q:v 2 {{.dir}}/{{.stem}}.mp4"
pipelines:
  broken:
    steps:
      - komp-2
      - ghost-action
logging:
  level: info
`

// emptyCmdYAML has an action with no cmd.
const emptyCmdYAML = `
actions:
  bad-action:
    cmd: ""
    when: "isVideo"
logging:
  level: info
`
