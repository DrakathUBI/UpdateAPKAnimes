Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$script:Macro = [ordered]@{
  name = 'macro-standalone'
  compatibility = 'v4_v5_unified'
  playback_profile = [ordered]@{
    speed = 1
    repetitions = 1
    post_playback_action = 'none'
    post_countdown_s = 5
  }
  actions = @()
}
$script:CurrentFile = ''
$script:SelectedIndex = -1
$script:Vars = @{}
$script:StopRequested = $false

function Write-Log($Text) {
  $txtLogs.AppendText("$Text`r`n")
  $txtLogs.SelectionStart = $txtLogs.Text.Length
  $txtLogs.ScrollToCaret()
}

function Expand-Vars([string]$Text) {
  if ([string]::IsNullOrEmpty($Text)) { return $Text }
  return [regex]::Replace($Text, '\$\{([A-Za-z0-9_]+)\}', {
    param($m)
    $n = $m.Groups[1].Value
    if ($script:Vars.ContainsKey($n)) { return [string]$script:Vars[$n] }
    return ''
  })
}

function Refresh-Grid {
  $grid.Rows.Clear()
  for ($i=0; $i -lt $script:Macro.actions.Count; $i++) {
    $a = $script:Macro.actions[$i]
    [void]$grid.Rows.Add(($i+1), $a.type, $a.label, $a.comment)
  }
}

function Select-Action([int]$Index) {
  if ($Index -lt 0 -or $Index -ge $script:Macro.actions.Count) { return }
  $script:SelectedIndex = $Index
  $a = $script:Macro.actions[$Index]
  $cmbType.Text = [string]$a.type
  $txtLabel.Text = [string]$a.label
  $txtComment.Text = [string]$a.comment
  $txtParams.Text = ($a.params | ConvertTo-Json -Depth 20)
}

function Build-LabelIndex {
  $labels = @{}
  for ($i=0; $i -lt $script:Macro.actions.Count; $i++) {
    $l = [string]$script:Macro.actions[$i].label
    if (-not [string]::IsNullOrWhiteSpace($l)) {
      if ($l -in @('Start','End','Next')) { throw "Label reservada: $l" }
      if ($labels.ContainsKey($l)) { throw "Label duplicada: $l" }
      $labels[$l] = $i
    }
  }
  return $labels
}

function Jump-ToLabel($Label, $Labels) {
  if ([string]::IsNullOrWhiteSpace($Label)) { return $null }
  if (-not $Labels.ContainsKey($Label)) { throw "Label inexistente: $Label" }
  return [int]$Labels[$Label]
}

function Compare-Values($Left, $Right, $Operator) {
  switch ($Operator) {
    'equals' { return [string]$Left -eq [string]$Right }
    'contains' { return ([string]$Left).Contains([string]$Right) }
    'regex' { return ([string]$Left -match [string]$Right) }
    'gt' { return [double]$Left -gt [double]$Right }
    'lt' { return [double]$Left -lt [double]$Right }
    'gte' { return [double]$Left -ge [double]$Right }
    'lte' { return [double]$Left -le [double]$Right }
    default { return $false }
  }
}

function Set-CursorPos([int]$x,[int]$y) {
  Add-Type @"
using System.Runtime.InteropServices;
public static class NativeWin {
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int X, int Y);
}
"@
  [NativeWin]::SetCursorPos($x,$y) | Out-Null
}

function Click-Mouse($Button) {
  Add-Type @"
using System.Runtime.InteropServices;
public static class NativeClick {
  [DllImport("user32.dll")] public static extern void mouse_event(int flags,int dx,int dy,int data,int extra);
}
"@
  switch ($Button) {
    'right' { $d=8; $u=16 }
    'middle' { $d=32; $u=64 }
    default { $d=2; $u=4 }
  }
  [NativeClick]::mouse_event($d,0,0,0,0)
  Start-Sleep -Milliseconds 20
  [NativeClick]::mouse_event($u,0,0,0,0)
}

function Scroll-Mouse([int]$Delta) {
  Add-Type @"
using System.Runtime.InteropServices;
public static class NativeScroll {
  [DllImport("user32.dll")] public static extern void mouse_event(int flags,int dx,int dy,int data,int extra);
}
"@
  [NativeScroll]::mouse_event(2048,0,0,$Delta,0)
}

function Save-Screenshot($Region, $OutputPath) {
  $x = [int]$Region[0]; $y = [int]$Region[1]; $w = [int]$Region[2]; $h = [int]$Region[3]
  $bmp = New-Object System.Drawing.Bitmap $w, $h
  $g = [System.Drawing.Graphics]::FromImage($bmp)
  $g.CopyFromScreen($x, $y, 0, 0, $bmp.Size)
  $bmp.Save($OutputPath, [System.Drawing.Imaging.ImageFormat]::Png)
  $g.Dispose(); $bmp.Dispose()
}

function Run-Action($Action, $Labels, [int]$Pc) {
  $p = $Action.params
  switch ($Action.type) {
    'set_variable' {
      $script:Vars[[string]$p.name] = Expand-Vars([string]$p.value)
      return $null
    }
    'create_variable' {
      $script:Vars[[string]$p.name] = [string]$p.default
      return $null
    }
    'text_output' {
      [System.Windows.Forms.SendKeys]::SendWait((Expand-Vars([string]$p.text)))
      return $null
    }
    'keypress' {
      [System.Windows.Forms.SendKeys]::SendWait((Expand-Vars([string]$p.key)))
      return $null
    }
    'hotkey' {
      $keys = [string]$p.keys
      $keys = $keys.Replace('CTRL','^').Replace('ALT','%').Replace('SHIFT','+').Replace('ENTER','{ENTER}')
      [System.Windows.Forms.SendKeys]::SendWait($keys)
      return $null
    }
    'window_focus' {
      try {
        $ws = New-Object -ComObject WScript.Shell
        $null = $ws.AppActivate((Expand-Vars([string]$p.window)))
      } catch { Write-Log "[WARN] foco de janela falhou: $($_.Exception.Message)" }
      return $null
    }
    'mouse_move' {
      Set-CursorPos ([int](Expand-Vars("$($p.x)"))) ([int](Expand-Vars("$($p.y)")))
      return $null
    }
    'mouse_click' {
      if ($p.x -ne $null -and $p.y -ne $null) {
        Set-CursorPos ([int](Expand-Vars("$($p.x)"))) ([int](Expand-Vars("$($p.y)")))
      }
      Click-Mouse ([string]$p.button)
      return $null
    }
    'mouse_scroll' {
      Scroll-Mouse ([int](Expand-Vars("$($p.delta)")))
      return $null
    }
    'wait_time' {
      $min = [int](Expand-Vars("$($p.minMs)")); if ($min -eq 0 -and $p.ms) { $min = [int]$p.ms }
      $max = [int](Expand-Vars("$($p.maxMs)")); if ($max -eq 0) { $max = $min }
      $ms = if ($max -gt $min) { Get-Random -Minimum $min -Maximum ($max+1) } else { $min }
      Start-Sleep -Milliseconds $ms
      return $null
    }
    'wait_until_time' {
      $target = [datetime](Expand-Vars([string]$p.isoTime))
      $now = Get-Date
      if ($target -gt $now) { Start-Sleep -Milliseconds ([int](($target-$now).TotalMilliseconds)) }
      return $null
    }
    'wait_for_file_event' {
      $file = Expand-Vars([string]$p.path)
      $timeout = [int]$p.timeoutMs; if ($timeout -le 0) { $timeout = 60000 }
      $start = Get-Date
      while (((Get-Date)-$start).TotalMilliseconds -lt $timeout) {
        if ($p.mode -eq 'exists' -and (Test-Path $file)) { return $null }
        Start-Sleep -Milliseconds 200
      }
      if ($p.onTimeout -eq 'goto') { return (Jump-ToLabel $p.timeoutLabel $Labels) }
      if ($p.onTimeout -eq 'abort') { throw 'Timeout no wait_for_file_event' }
      return $null
    }
    'capture_bitmap' {
      $out = if ($p.outputPath) { Expand-Vars([string]$p.outputPath) } else { Join-Path $env:TEMP ("macro_capture_{0}.png" -f [DateTimeOffset]::Now.ToUnixTimeMilliseconds()) }
      $r = if ($p.region) { $p.region } else { @(0,0,300,200) }
      Save-Screenshot $r $out
      if ($p.outputVar) { $script:Vars[[string]$p.outputVar] = $out }
      return $null
    }
    'save_variable' {
      $v = [string]$script:Vars[[string]$p.name]
      if ($p.destination -eq 'clipboard') {
        Set-Clipboard -Value $v
      } else {
        $target = Expand-Vars([string]$p.path)
        $dir = Split-Path -Path $target -Parent
        if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
        if ($p.append) { Add-Content -Path $target -Value $v } else { Set-Content -Path $target -Value $v }
      }
      return $null
    }
    'data_list' {
      $key = "__datalist_idx_$($p.name)"
      $idx = if ($script:Vars.ContainsKey($key)) { [int]$script:Vars[$key] } else { 0 }
      $arr = @()
      if ($p.file) {
        $arr = Get-Content -Path (Expand-Vars([string]$p.file)) -ErrorAction SilentlyContinue
      } elseif ($p.items) {
        $arr = @($p.items)
      }
      $cur = if ($idx -lt $arr.Count) { [string]$arr[$idx] } else { '' }
      $script:Vars[[string]($p.target ? $p.target : $p.name)] = $cur
      $script:Vars[$key] = ($idx + 1)
      $script:Vars["__datalist_end_$($p.name)"] = (($idx + 1) -ge $arr.Count)
      return $null
    }
    'calculation' {
      $expr = Expand-Vars([string]$p.expression)
      $expr = $expr.Replace('^','**').Replace('pi',[math]::PI).Replace('e',[math]::E)
      $res = Invoke-Expression $expr
      $script:Vars[[string]$p.target] = $res
      return $null
    }
    'if_then_else' {
      $ok = Compare-Values (Expand-Vars("$($p.left)")) (Expand-Vars("$($p.right)")) ([string]$p.operator)
      if ($ok) { return (Jump-ToLabel $p.thenLabel $Labels) }
      return (Jump-ToLabel $p.elseLabel $Labels)
    }
    'goto' { return (Jump-ToLabel $p.label $Labels) }
    'repeat' {
      $rk = "__repeat_$Pc"
      $c = if ($script:Vars.ContainsKey($rk)) { [int]$script:Vars[$rk] + 1 } else { 1 }
      $script:Vars[$rk] = $c
      if ($p.count -and $c -le [int]$p.count) { return (Jump-ToLabel $p.label $Labels) }
      return (Jump-ToLabel $p.afterLabel $Labels)
    }
    'show_notification' {
      Write-Log "[NOTIFY] $(Expand-Vars([string]$p.text))"
      return $null
    }
    'show_message_box' {
      $txt = Expand-Vars([string]$p.text)
      [void][System.Windows.Forms.MessageBox]::Show($txt, 'Macro Recorder')
      return $null
    }
    'beep' {
      [console]::beep(1200,130)
      return $null
    }
    'execute_program' {
      if ($p.program) { Start-Process -FilePath (Expand-Vars([string]$p.program)) -ArgumentList @($p.args) }
      return $null
    }
    'stop' {
      $script:StopRequested = $true
      return $null
    }
    default {
      Write-Log "[WARN] ação '$($Action.type)' executada em modo best-effort"
      if ($p.outputVarX) { $script:Vars[[string]$p.outputVarX] = 0 }
      if ($p.outputVarY) { $script:Vars[[string]$p.outputVarY] = 0 }
      if ($p.outputVar) { $script:Vars[[string]$p.outputVar] = '' }
      return $null
    }
  }
}

function Run-Macro {
  try {
    $script:StopRequested = $false
    $labels = Build-LabelIndex
    $pc = 0
    Write-Log "Compat mode: $($script:Macro.compatibility)"

    while ($pc -lt $script:Macro.actions.Count -and -not $script:StopRequested) {
      $a = $script:Macro.actions[$pc]
      if ($a.enabled -eq $false) { $pc++; continue }
      Write-Log ("[{0}] {1}" -f ($pc+1), $a.type)
      $next = Run-Action $a $labels $pc
      if ($next -ne $null) { $pc = [int]$next } else { $pc++ }
      [System.Windows.Forms.Application]::DoEvents()
    }

    $outDir = Join-Path $env:TEMP 'macro-recorder-no-admin'
    if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Force -Path $outDir | Out-Null }
    $outFile = Join-Path $outDir 'last-vars.json'
    ($script:Vars | ConvertTo-Json -Depth 20) | Set-Content -Path $outFile
    Write-Log "Concluído. Vars em: $outFile"
  } catch {
    Write-Log "[ERRO] $($_.Exception.Message)"
  }
}

# ===== UI =====
$form = New-Object System.Windows.Forms.Form
$form.Text = 'Macro Recorder Standalone (Sem npm / sem Python)'
$form.StartPosition = 'CenterScreen'
$form.Size = New-Object System.Drawing.Size(1450, 920)
$form.BackColor = [System.Drawing.Color]::FromArgb(10, 16, 32)
$form.ForeColor = [System.Drawing.Color]::White

$topPanel = New-Object System.Windows.Forms.Panel
$topPanel.Dock = 'Top'; $topPanel.Height = 64
$topPanel.BackColor = [System.Drawing.Color]::FromArgb(23, 31, 55)
$form.Controls.Add($topPanel)

$btnOpen = New-Object System.Windows.Forms.Button
$btnOpen.Text = 'Abrir'; $btnOpen.Location = New-Object Drawing.Point(15,16); $btnOpen.Size = New-Object Drawing.Size(90,32)
$btnSave = New-Object System.Windows.Forms.Button
$btnSave.Text = 'Salvar'; $btnSave.Location = New-Object Drawing.Point(110,16); $btnSave.Size = New-Object Drawing.Size(90,32)
$btnAdd = New-Object System.Windows.Forms.Button
$btnAdd.Text = '+ Ação'; $btnAdd.Location = New-Object Drawing.Point(205,16); $btnAdd.Size = New-Object Drawing.Size(90,32)
$btnRun = New-Object System.Windows.Forms.Button
$btnRun.Text = '▶ Executar'; $btnRun.Location = New-Object Drawing.Point(300,16); $btnRun.Size = New-Object Drawing.Size(110,32)
$btnRun.BackColor = [System.Drawing.Color]::FromArgb(24,180,110)
$lblTitle = New-Object System.Windows.Forms.Label
$lblTitle.Text = 'Macro Recorder All-in-One • Windows Standalone'
$lblTitle.AutoSize = $true; $lblTitle.Location = New-Object Drawing.Point(430,22)

$topPanel.Controls.AddRange(@($btnOpen,$btnSave,$btnAdd,$btnRun,$lblTitle))

$tabFake = New-Object System.Windows.Forms.Panel
$tabFake.Dock = 'Top'; $tabFake.Height = 40; $tabFake.BackColor = [System.Drawing.Color]::FromArgb(28,38,70)
$form.Controls.Add($tabFake)
$lblTabs = New-Object System.Windows.Forms.Label
$lblTabs.Text = 'File  |  Record and Edit  |  Playback  |  Help    (UI única e standalone)'
$lblTabs.AutoSize = $true; $lblTabs.Location = New-Object Drawing.Point(20,11)
$tabFake.Controls.Add($lblTabs)

$left = New-Object System.Windows.Forms.GroupBox
$left.Text = 'Script (ações)'; $left.Location = New-Object Drawing.Point(12,110); $left.Size = New-Object Drawing.Size(650,760)
$mid = New-Object System.Windows.Forms.GroupBox
$mid.Text = 'Editor de ação'; $mid.Location = New-Object Drawing.Point(670,110); $mid.Size = New-Object Drawing.Size(360,760)
$right = New-Object System.Windows.Forms.GroupBox
$right.Text = 'Console / Playback'; $right.Location = New-Object Drawing.Point(1038,110); $right.Size = New-Object Drawing.Size(390,760)
$form.Controls.AddRange(@($left,$mid,$right))

$grid = New-Object System.Windows.Forms.DataGridView
$grid.Location = New-Object Drawing.Point(14,24); $grid.Size = New-Object Drawing.Size(620,720)
$grid.ReadOnly = $true; $grid.AllowUserToAddRows = $false; $grid.SelectionMode = 'FullRowSelect'; $grid.MultiSelect = $false
[void]$grid.Columns.Add('idx','#')
[void]$grid.Columns.Add('type','Type')
[void]$grid.Columns.Add('label','Label')
[void]$grid.Columns.Add('comment','Comment')
$left.Controls.Add($grid)

$lblType = New-Object Windows.Forms.Label; $lblType.Text='Tipo'; $lblType.Location=New-Object Drawing.Point(16,35)
$cmbType = New-Object Windows.Forms.ComboBox; $cmbType.Location=New-Object Drawing.Point(16,55); $cmbType.Size=New-Object Drawing.Size(320,24)
$types = @('set_variable','create_variable','text_output','keypress','hotkey','mouse_move','mouse_click','mouse_scroll','smart_click','window_focus','execute_program','embed_macro','wait_time','wait_until_time','wait_for_file_event','wait_for_hotkey','wait_for_text_input','wait_for_pixel_color','wait_for_desktop_change','find_image','find_text_ocr','capture_bitmap','capture_text_ocr','capture_barcode_qr','scrape_webpage','save_variable','data_list','calculation','if_then_else','goto','repeat','show_notification','show_message_box','beep','phrase_insertion','ai_object_search','stop')
$cmbType.Items.AddRange($types)
$lblLabel = New-Object Windows.Forms.Label; $lblLabel.Text='Label'; $lblLabel.Location=New-Object Drawing.Point(16,90)
$txtLabel = New-Object Windows.Forms.TextBox; $txtLabel.Location=New-Object Drawing.Point(16,110); $txtLabel.Size=New-Object Drawing.Size(320,24)
$lblComment = New-Object Windows.Forms.Label; $lblComment.Text='Comment'; $lblComment.Location=New-Object Drawing.Point(16,145)
$txtComment = New-Object Windows.Forms.TextBox; $txtComment.Location=New-Object Drawing.Point(16,165); $txtComment.Size=New-Object Drawing.Size(320,24)
$lblParams = New-Object Windows.Forms.Label; $lblParams.Text='Params (JSON)'; $lblParams.Location=New-Object Drawing.Point(16,200)
$txtParams = New-Object Windows.Forms.TextBox; $txtParams.Location=New-Object Drawing.Point(16,220); $txtParams.Size=New-Object Drawing.Size(320,460); $txtParams.Multiline=$true; $txtParams.ScrollBars='Vertical'
$btnApply = New-Object Windows.Forms.Button; $btnApply.Text='Aplicar'; $btnApply.Location=New-Object Drawing.Point(16,690); $btnApply.Size=New-Object Drawing.Size(150,30)
$btnDel = New-Object Windows.Forms.Button; $btnDel.Text='Excluir'; $btnDel.Location=New-Object Drawing.Point(186,690); $btnDel.Size=New-Object Drawing.Size(150,30)
$mid.Controls.AddRange(@($lblType,$cmbType,$lblLabel,$txtLabel,$lblComment,$txtComment,$lblParams,$txtParams,$btnApply,$btnDel))

$lblStart = New-Object Windows.Forms.Label; $lblStart.Text='Start at label'; $lblStart.Location=New-Object Drawing.Point(14,28)
$txtStart = New-Object Windows.Forms.TextBox; $txtStart.Location=New-Object Drawing.Point(14,48); $txtStart.Size=New-Object Drawing.Size(360,24)
$txtLogs = New-Object Windows.Forms.TextBox
$txtLogs.Location = New-Object Drawing.Point(14,82); $txtLogs.Size = New-Object Drawing.Size(360,660)
$txtLogs.Multiline = $true; $txtLogs.ScrollBars = 'Vertical'; $txtLogs.ReadOnly = $true
$right.Controls.AddRange(@($lblStart,$txtStart,$txtLogs))

$grid.add_CellClick({
  if ($_.RowIndex -ge 0) { Select-Action $_.RowIndex }
})

$btnAdd.Add_Click({
  $script:Macro.actions += [pscustomobject]@{ type='set_variable'; label=$null; comment=$null; params=@{} }
  Refresh-Grid
  Select-Action ($script:Macro.actions.Count - 1)
})

$btnApply.Add_Click({
  if ($script:SelectedIndex -lt 0) { return }
  try {
    $paramsObj = @{}
    if (-not [string]::IsNullOrWhiteSpace($txtParams.Text)) {
      $paramsObj = $txtParams.Text | ConvertFrom-Json -AsHashtable
    }
    $script:Macro.actions[$script:SelectedIndex] = [pscustomobject]@{
      type = $cmbType.Text
      label = $(if ($txtLabel.Text) { $txtLabel.Text } else { $null })
      comment = $(if ($txtComment.Text) { $txtComment.Text } else { $null })
      params = $paramsObj
    }
    Refresh-Grid
    Write-Log 'Ação atualizada.'
  } catch {
    Write-Log "[ERRO] JSON inválido: $($_.Exception.Message)"
  }
})

$btnDel.Add_Click({
  if ($script:SelectedIndex -lt 0) { return }
  $list = New-Object System.Collections.ArrayList
  foreach ($a in $script:Macro.actions) { [void]$list.Add($a) }
  $list.RemoveAt($script:SelectedIndex)
  $script:Macro.actions = @($list)
  $script:SelectedIndex = -1
  Refresh-Grid
})

$btnOpen.Add_Click({
  $dlg = New-Object System.Windows.Forms.OpenFileDialog
  $dlg.Filter = 'Macro JSON (*.json)|*.json'
  if ($dlg.ShowDialog() -eq 'OK') {
    try {
      $script:CurrentFile = $dlg.FileName
      $raw = Get-Content -Path $script:CurrentFile -Raw
      $script:Macro = $raw | ConvertFrom-Json -AsHashtable
      if (-not $script:Macro.actions) { $script:Macro.actions = @() }
      Refresh-Grid
      Write-Log "Arquivo aberto: $script:CurrentFile"
    } catch {
      Write-Log "[ERRO] Não foi possível abrir: $($_.Exception.Message)"
    }
  }
})

$btnSave.Add_Click({
  $dlg = New-Object System.Windows.Forms.SaveFileDialog
  $dlg.Filter = 'Macro JSON (*.json)|*.json'
  if ($script:CurrentFile) { $dlg.FileName = [IO.Path]::GetFileName($script:CurrentFile) }
  if ($dlg.ShowDialog() -eq 'OK') {
    try {
      $script:CurrentFile = $dlg.FileName
      ($script:Macro | ConvertTo-Json -Depth 30) | Set-Content -Path $script:CurrentFile
      Write-Log "Arquivo salvo: $script:CurrentFile"
    } catch {
      Write-Log "[ERRO] Não foi possível salvar: $($_.Exception.Message)"
    }
  }
})

$btnRun.Add_Click({
  try {
    Write-Log 'Executando macro...'
    if ($txtStart.Text) {
      $labels = Build-LabelIndex
      if ($labels.ContainsKey($txtStart.Text)) {
        $start = [int]$labels[$txtStart.Text]
        if ($start -gt 0) {
          $new = @()
          for ($i=$start; $i -lt $script:Macro.actions.Count; $i++) { $new += $script:Macro.actions[$i] }
          $bak = $script:Macro.actions
          $script:Macro.actions = $new
          Run-Macro
          $script:Macro.actions = $bak
          return
        }
      }
    }
    Run-Macro
  } catch {
    Write-Log "[ERRO] execução: $($_.Exception.Message)"
  }
})

Write-Log 'Pronto. Use Abrir/Salvar/+Ação/Executar.'
[void]$form.ShowDialog()
