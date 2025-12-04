using GreenerBlazor.Models;
using Microsoft.FluentUI.AspNetCore.Components;

namespace GreenerBlazor.Helpers;

public class TestcaseRow
{
    public TestcaseRow(string id, string sessionId, string name, TestcaseStatus status)
    {
        Id = id;
        Session = sessionId;
        Name = name;
        (StatusTitle, StatusIcon) = Ext.ConvertTestcaseStatus(status);
    }

    public TestcaseRow(string id, string sessionId, string name, string status)
    {
        Id = id;
        Session = sessionId;
        Name = name;
        (StatusTitle, StatusIcon) = Ext.ConvertTestcaseStatus(status);
    }

    public string Id { get; }
    public string Session { get; }
    public string Name { get; }
    public string StatusTitle { get; }
    public Icon StatusIcon { get; }
}
