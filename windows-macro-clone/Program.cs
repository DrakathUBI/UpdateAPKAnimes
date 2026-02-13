using System.Data;
using System.Diagnostics;
using System.Text.Json;
using System.Text.Json.Serialization;

var argsList = args.ToList();
if (argsList.Contains("--help") || args.Length == 0)
{
    Console.WriteLine("MacroClone (sem admin)");
    Console.WriteLine("Uso:");
    Console.WriteLine("  MacroClone.exe --play macro.json [--silent] [--start-at Label]");
    return;
}

string? macroPath = GetArgValue(argsList, "--play");
bool silent = argsList.Contains("--silent");
string? startAt = GetArgValue(argsList, "--start-at");

if (string.IsNullOrWhiteSpace(macroPath) || !File.Exists(macroPath))
{
    Console.Error.WriteLine("Arquivo de macro não encontrado.");
    return;
}

var macro = JsonSerializer.Deserialize<MacroScript>(File.ReadAllText(macroPath), JsonOptions.Default);
if (macro is null)
{
    Console.Error.WriteLine("Falha ao carregar macro.");
    return;
}

var executor = new MacroExecutor(silent);
await executor.RunAsync(macro, startAt);

static string? GetArgValue(List<string> args, string key)
{
    var idx = args.IndexOf(key);
    if (idx >= 0 && idx + 1 < args.Count) return args[idx + 1];
    return null;
}

public static class JsonOptions
{
    public static JsonSerializerOptions Default => new()
    {
        PropertyNameCaseInsensitive = true,
        WriteIndented = true,
        Converters = { new JsonStringEnumConverter() }
    };
}

public sealed class MacroExecutor(bool silent)
{
    private readonly Dictionary<string, object> _vars = new(StringComparer.OrdinalIgnoreCase);

    public async Task RunAsync(MacroScript macro, string? startAt)
    {
        var actions = macro.Actions.Where(a => a.Enabled).ToList();
        var labels = actions
            .Where(a => !string.IsNullOrWhiteSpace(a.Label))
            .Select((a, i) => (a.Label!, i))
            .ToDictionary(x => x.Item1, x => x.i, StringComparer.OrdinalIgnoreCase);

        int pointer = 0;
        if (!string.IsNullOrWhiteSpace(startAt) && labels.TryGetValue(startAt, out var jump)) pointer = jump;

        while (pointer < actions.Count)
        {
            var a = actions[pointer];
            if (!silent) Console.WriteLine($"[{pointer + 1}] {a.Type} {(a.Comment ?? "")}");

            switch (a.Type)
            {
                case ActionType.WaitTime:
                    await Task.Delay(a.WaitMs ?? 1000);
                    pointer++;
                    break;

                case ActionType.SetVariable:
                    if (!string.IsNullOrWhiteSpace(a.VariableName)) _vars[a.VariableName] = Expand(a.Value ?? string.Empty);
                    pointer++;
                    break;

                case ActionType.Calculation:
                    if (!string.IsNullOrWhiteSpace(a.VariableName) && !string.IsNullOrWhiteSpace(a.Expression))
                    {
                        var expr = Expand(a.Expression);
                        var dt = new DataTable();
                        var result = dt.Compute(expr, "");
                        _vars[a.VariableName] = Convert.ToString(result) ?? "0";
                    }
                    pointer++;
                    break;

                case ActionType.IfThenElse:
                    {
                        var left = Expand(a.Left ?? "");
                        var right = Expand(a.Right ?? "");
                        var ok = a.Operator switch
                        {
                            "equals" => string.Equals(left, right, StringComparison.OrdinalIgnoreCase),
                            "contains" => left.Contains(right, StringComparison.OrdinalIgnoreCase),
                            "regex" => System.Text.RegularExpressions.Regex.IsMatch(left, right),
                            ">" => decimal.Parse(left) > decimal.Parse(right),
                            "<" => decimal.Parse(left) < decimal.Parse(right),
                            _ => false
                        };

                        var target = ok ? a.ThenLabel : a.ElseLabel;
                        if (target is not null && labels.TryGetValue(target, out var p2)) pointer = p2;
                        else pointer++;
                    }
                    break;

                case ActionType.Goto:
                    if (a.TargetLabel is not null && labels.TryGetValue(a.TargetLabel, out var p)) pointer = p;
                    else throw new InvalidOperationException($"Label não encontrada: {a.TargetLabel}");
                    break;

                case ActionType.TextOutput:
                    Console.WriteLine(Expand(a.Value ?? ""));
                    pointer++;
                    break;

                case ActionType.ExecuteProgram:
                    if (!string.IsNullOrWhiteSpace(a.FilePath))
                    {
                        Process.Start(new ProcessStartInfo
                        {
                            FileName = Expand(a.FilePath),
                            Arguments = Expand(a.Arguments ?? ""),
                            UseShellExecute = true
                        });
                    }
                    pointer++;
                    break;

                case ActionType.SaveVariable:
                    if (!string.IsNullOrWhiteSpace(a.VariableName) && _vars.TryGetValue(a.VariableName, out var value) && !string.IsNullOrWhiteSpace(a.FilePath))
                    {
                        var path = Expand(a.FilePath);
                        var text = Convert.ToString(value) ?? string.Empty;
                        if (a.Append) File.AppendAllText(path, text + Environment.NewLine);
                        else File.WriteAllText(path, text);
                    }
                    pointer++;
                    break;

                case ActionType.Beep:
                    Console.Beep();
                    pointer++;
                    break;

                default:
                    Console.WriteLine($"[AVISO] Ação {a.Type} requer integração de sistema/UI e está em modo seguro sem admin.");
                    pointer++;
                    break;
            }
        }
    }

    private string Expand(string input)
    {
        var output = input;
        foreach (var kv in _vars)
        {
            output = output.Replace("${" + kv.Key + "}", Convert.ToString(kv.Value));
        }
        return output;
    }
}

public sealed class MacroScript
{
    public string Name { get; set; } = "Nova Macro";
    public List<ActionItem> Actions { get; set; } = [];
}

public sealed class ActionItem
{
    public ActionType Type { get; set; }
    public bool Enabled { get; set; } = true;
    public string? Label { get; set; }
    public string? Comment { get; set; }

    public int? WaitMs { get; set; }
    public string? VariableName { get; set; }
    public string? Value { get; set; }
    public string? Expression { get; set; }
    public string? TargetLabel { get; set; }
    public string? ThenLabel { get; set; }
    public string? ElseLabel { get; set; }
    public string? Operator { get; set; }
    public string? Left { get; set; }
    public string? Right { get; set; }
    public string? FilePath { get; set; }
    public string? Arguments { get; set; }
    public bool Append { get; set; }
}

public enum ActionType
{
    MouseClick,
    SmartClick,
    MouseMove,
    MouseScroll,
    KeyPress,
    Hotkey,
    TextOutput,
    WaitTime,
    WaitUntilTime,
    WaitForHotkey,
    WaitForTextInput,
    WaitForFileEvent,
    WaitPixelColor,
    WaitDesktopChange,
    FindImage,
    FindTextOcr,
    CaptureBitmap,
    CaptureTextOcr,
    CaptureBarcodeQr,
    ScrapeWebpage,
    SetVariable,
    Calculation,
    SaveVariable,
    DataList,
    Goto,
    Repeat,
    IfThenElse,
    WindowFocus,
    ExecuteProgram,
    EmbedMacro,
    ShowNotification,
    ShowMessageBox,
    Beep
}
