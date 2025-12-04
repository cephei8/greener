namespace GreenerBlazor.Models;

public class LabelDto
{
    public required string Key { get; set; }
    public string? Value { get; set; }
    public required DateTime CreatedAt { get; set; }
}
