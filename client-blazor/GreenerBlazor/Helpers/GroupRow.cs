using System.Text.Json;
using Microsoft.FluentUI.AspNetCore.Components;

namespace GreenerBlazor.Helpers;

public class GroupRow
{
    public GroupRow(
        string id,
        string status,
        IReadOnlyCollection<string> header,
        IReadOnlyCollection<string?> columns
    )
    {
        Id = id;
        (StatusTitle, StatusIcon) = Ext.ConvertTestcaseStatus(status);
        Header = header;
        Columns = columns;
    }

    public string Id { get; }
    public string StatusTitle { get; }
    public Icon StatusIcon { get; }
    public IReadOnlyCollection<string> Header { get; }
    public IReadOnlyCollection<string?> Columns { get; }

    public string CreateGroupParameterJson()
    {
        var groupData = new[] { [.. Header], Columns.ToArray() };
        var json = JsonSerializer.Serialize(groupData);
        return Uri.EscapeDataString(json);
    }
}
