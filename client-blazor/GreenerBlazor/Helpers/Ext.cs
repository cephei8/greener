using System.Text.Json;
using GreenerBlazor.Models;
using Microsoft.FluentUI.AspNetCore.Components;
using Icons = Microsoft.FluentUI.AspNetCore.Components.Icons;

namespace GreenerBlazor.Helpers;

public static class Ext
{
    private static readonly Icon StatusPass = new Icons.Color.Size20.CheckmarkCircle();
    private static readonly Icon StatusFail = new Icons.Color.Size20.DismissCircle();
    private static readonly Icon StatusErr = new Icons.Color.Size20.DismissCircle();
    private static readonly Icon StatusSkip = new Icons.Filled.Size20.SkipForwardTab();
    private static readonly Icon StatusNa = new Icons.Filled.Size20.Question();

    public static (string, Icon) ConvertTestcaseStatus(TestcaseStatus? status)
    {
        return status switch
        {
            TestcaseStatus.Pass => ("Passed", StatusPass),
            TestcaseStatus.Fail => ("Failed", StatusFail),
            TestcaseStatus.Error => ("Error", StatusErr),
            TestcaseStatus.Skip => ("Skipped", StatusSkip),
            null => ("n/a", StatusNa),
            _ => ("Unknown", StatusNa),
        };
    }

    public static (string, Icon) ConvertTestcaseStatus(string? status)
    {
        if (string.IsNullOrEmpty(status))
        {
            return ("n/a", StatusNa);
        }

        return status.ToLowerInvariant() switch
        {
            "pass" => ("Passed", StatusPass),
            "fail" => ("Failed", StatusFail),
            "error" => ("Error", StatusErr),
            "skip" => ("Skipped", StatusSkip),
            _ => ("Unknown", StatusNa),
        };
    }

    public static string GroupIdJsonToString(
        JsonDocument idJson,
        IReadOnlyCollection<string> header
    )
    {
        List<string> keys = [];

        foreach (var h in header)
        {
            if (idJson.RootElement.TryGetProperty(h, out var prop))
            {
                var value = prop.ValueKind switch
                {
                    JsonValueKind.String => prop.GetString(),
                    JsonValueKind.Number => prop.GetDecimal().ToString(),
                    JsonValueKind.True => "true",
                    JsonValueKind.False => "false",
                    JsonValueKind.Null => "null",
                    _ => prop.GetRawText(),
                };

                keys.Add(value ?? "null");
            }
            else
            {
                keys.Add("null");
            }
        }

        return string.Join(", ", keys);
    }
}
