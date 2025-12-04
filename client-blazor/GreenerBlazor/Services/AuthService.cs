using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class AuthService(ApiClient apiClient, AppAuthenticationStateProvider authStateProvider)
{
    public async Task LoginAsync(LoginRequestDto requestDto, CancellationToken cancellationToken)
    {
        var loginResponse = await apiClient.PostAsync<LoginRequestDto, TokenResponseDto>(
            "/api/v1/auth/login",
            requestDto,
            cancellationToken
        );

        await authStateProvider.SetStateAsync(
            new UserSession
            {
                Username = requestDto.Username,
                AccessToken = loginResponse.AccessToken,
                AccessTokenExpiresAt = loginResponse.AccessTokenExpiresAt,
                RefreshToken = loginResponse.RefreshToken,
                RefreshTokenExpiresAt = loginResponse.RefreshTokenExpiresAt,
            }
        );
    }

    public async Task RefreshAsync(
        RefreshRequestDto requestDto,
        CancellationToken cancellationToken
    )
    {
        var tokenDto = await apiClient.PostAsync<RefreshRequestDto, TokenResponseDto>(
            "/api/v1/auth/refresh",
            requestDto,
            cancellationToken
        );

        var userSession =
            await authStateProvider.GetUserSessionAsync()
            ?? throw new InvalidOperationException("No user session available for token refresh.");

        await authStateProvider.SetStateAsync(
            new UserSession
            {
                Username = userSession.Username,
                AccessToken = tokenDto.AccessToken,
                AccessTokenExpiresAt = tokenDto.AccessTokenExpiresAt,
                RefreshToken = tokenDto.RefreshToken,
                RefreshTokenExpiresAt = tokenDto.RefreshTokenExpiresAt,
            }
        );
    }

    public async Task ChangePasswordAsync(
        ChangePasswordRequestDto request,
        CancellationToken cancellationToken
    )
    {
        await apiClient.PostAsync("/api/v1/auth/change-password", request, cancellationToken);
    }

    public async Task LogoutAsync(CancellationToken cancellationToken)
    {
        await apiClient.PostAsync("/api/v1/auth/logout", new { }, cancellationToken);
    }
}
