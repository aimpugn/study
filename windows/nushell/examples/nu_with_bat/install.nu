# 다음과 같은 구조로 패키징되어 있다고 가정합니다.
# ├─assets
# │  ├─configs
# │  └─themes
# ├─bin
# └─programs
#     ├─nu
#     └─terminal
#
# C:\Users\username\dev\tools 디렉토리 하위로 설치합니다.
#
# 📌 경로 설정
let home = $env.USERPROFILE
let toolsDir = [ $home "dev" "tools" ] | path join

# Assets 디렉토리
let assetsDir = [ $toolsDir "assets" ] | path join
let configDir = [ $assetsDir "configs" ] | path join
let defaultNuConfig = [ $configDir "config.nu" ] | path join
let defaultNuEnv = [ $configDir "env.nu" ] | path join
let defaultTerminalSettings = [ $configDir "terminal.settings.json" ] | path join

# bin 디렉토리
let binDir = [ $toolsDir "bin" ] | path join

# programs 디렉토리
let programsDir = [ $toolsDir "programs" ] | path join
let nuDir = [ $programsDir "nu" ] | path join
let terminalDir = [ $programsDir "terminal" ] | path join

# Windows Terminal 설정 파일 경로 찾기 (안전한 검색)
let terminalPath = [ $env.LOCALAPPDATA "Microsoft" "Windows Terminal" ] | path join
let terminalConfig = if $terminalPath != "" {
    [ $terminalPath "settings.json" ] | path join
} else {
    print "⚠ Windows Terminal 설정 파일을 찾을 수 없습니다."; ""
}

# 📌 디렉토리 생성 (존재 여부 확인 후 생성)
print "📂 Create Directories..."
for dir in [$toolsDir, $binDir, $programsDir, $assetsDir, $nuDir, $terminalDir] {
    if not ($dir | path exists) {
        mkdir $dir
        print $"✅ Created: ($dir)"
    } else {
        print $"🔹 Already exists: ($dir)"
    }
}

# 📌 파일 복사 (존재 여부 확인 후 실행)
print $"📂 Copying assets, bin, programs into ($toolsDir)..."
for dir in ["assets", "bin", "programs"] {
    let src = $"./($dir)"
    if ($src | path exists) {
        cp -r $src $toolsDir
        print $"✅ Copied: ($src) -> ($toolsDir)"
    } else {
        print $"⚠ Skipping: ($src) does not exist"
    }
}

# 📌 파일 백업 및 덮어쓰기 함수
def backup_and_overwrite [file_path: string, backup_suffix: string, source_file: string] {
    if ($file_path | path exists) {
        let backup_path = $"($file_path).($backup_suffix)"
        cp $file_path $backup_path
        print $"📂 Backup created: ($backup_path)"
    } else {
        print $"⚠ No existing file to backup: ($file_path)"
    }

    if ($source_file | path exists) {
        print $"✍ Overwriting: ($file_path) -> ($source_file)"
        open $source_file | save -f $file_path
    } else {
        print $"⚠ Source file not found: ($source_file)"
    }
}

# 📌 백업 및 덮어쓰기 실행
let backupSuffix = (date now | format date "%Y%m%d_%H%M%S")

print "🔄 Backing up and Overwriting Nushell configs..."
backup_and_overwrite $nu.config-path $backupSuffix $defaultNuConfig
backup_and_overwrite $nu.env-path $backupSuffix $defaultNuEnv

if $terminalConfig != "" {
    print "🔄 Backing up and Overwriting Windows Terminal settings..."
    backup_and_overwrite $terminalConfig $backupSuffix $defaultTerminalSettings
} else {
    print "⚠ Windows Terminal 설정 파일이 존재하지 않아 스킵합니다."
}

print "Done"
