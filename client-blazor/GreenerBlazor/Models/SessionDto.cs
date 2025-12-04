using System.Text.Json;

namespace GreenerBlazor.Models;

public class SessionDto
{
    public required string Id { get; set; }
    public string? Description { get; set; }
    public JsonDocument? Baggage { get; set; }
    public required DateTime CreatedAt { get; set; }
}
