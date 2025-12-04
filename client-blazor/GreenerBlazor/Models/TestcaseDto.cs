using System.Text.Json;

namespace GreenerBlazor.Models;

public class TestcaseDto
{
    public required string Id { get; set; }
    public required string SessionId { get; set; }
    public required string Name { get; set; }
    public string? Classname { get; set; }
    public string? File { get; set; }
    public string? Testsuite { get; set; }
    public required TestcaseStatus Status { get; set; }
    public string? Output { get; set; }
    public JsonDocument? Baggage { get; set; }
    public required DateTime CreatedAt { get; set; }
}
