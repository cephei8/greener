namespace GreenerBlazor.Models;

public readonly struct UserSession
{
    public required string Username { get; init; }
    public required string AccessToken { get; init; }
    public required DateTime AccessTokenExpiresAt { get; init; }
    public required string RefreshToken { get; init; }
    public required DateTime RefreshTokenExpiresAt { get; init; }
}
