; Inno Setup Script for Fyrnox for Windows
#define MyAppName "Fyrnox"
#ifndef MyAppVersion
  #define VersionFile FileOpen("VERSION")
  #define MyAppVersion Trim(FileRead(VersionFile))
  #expr FileClose(VersionFile)
#endif
#define MyAppPublisher "Powered by Yousef Kia"
#define MyAppURL "https://github.com/yousefkiamozafari"
#define MyAppExeName "Fyrnox.exe"

[Setup]
AppId={{E4A584EF-F1A1-440A-AB06-38E888100028}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DisableProgramGroupPage=yes
OutputDir=build\dist
OutputBaseFilename=Fyrnox-InnoSetup-v{#MyAppVersion}
SetupIconFile=icon.ico
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Files]
Source: "build\Fyrnox-x64.exe"; DestDir: "{app}"; DestName: "Fyrnox.exe"; Flags: ignoreversion
Source: "icon.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "README.md"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\icon.ico"; Tasks: desktopicon

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(MyAppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent
