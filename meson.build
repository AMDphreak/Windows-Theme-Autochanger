project(
  'windows-theme-autochanger',
  'cpp', # Use a supported language placeholder, since Go is not supported natively by meson
  version : '0.1',
  default_options : ['warning_level=3']
)

# Find Go compiler
go = find_program('go')

# Build the GUI executable
gui = custom_target(
  'gui',
  output : 'windows-theme-autochanger.exe',
  command : [go, 'build', '-C', meson.project_source_root(), '-ldflags=-H windowsgui', '-o', '@OUTPUT@'],
  build_by_default : true
)

# Build the service executable
service = custom_target(
  'service',
  output : 'windows-theme-autochanger-service.exe',
  command : [go, 'build', '-C', meson.project_source_root(), '-o', '@OUTPUT@', './cmd/service'],
  install : true,
  install_dir : './cmd/service',
  build_by_default : true
)