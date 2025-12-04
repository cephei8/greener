namespace GreenerBlazor.Models;

public class ApiKeyDto
{
    public required string Id { get; set; }
    public string? Description { get; set; }
    public required DateTime CreatedAt { get; set; }
}

public class CreateApiKeyRequestDto
{
    public string? Description { get; set; }
}

public class CreateApiKeyResponseDto
{
    public required string Id { get; set; }
    public string? Description { get; set; }
    public required string Key { get; set; }
    public required DateTime CreatedAt { get; set; }
}
