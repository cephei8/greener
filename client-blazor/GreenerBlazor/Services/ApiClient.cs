using System.Text;
using System.Text.Json;
using GreenerBlazor.Models;
using Microsoft.AspNetCore.Components;

namespace GreenerBlazor.Services;

public class ApiClient(
    HttpClient httpClient,
    UserSessionService userSessionService,
    AppAuthenticationStateProvider authStateProvider,
    NavigationManager navigationManager
)
{
    public async Task<TResponse> PostAsync<TRequest, TResponse>(
        string endpoint,
        TRequest request,
        CancellationToken cancellationToken
    )
    {
        var json = JsonSerializer.Serialize(request, AppJsonSerializerContext.Default.Options);
        var content = new StringContent(json, Encoding.UTF8, "application/json");

        await AddAuthorizationHeader();

        var response = await httpClient.PostAsync(endpoint, content, cancellationToken);
        await EnsureSuccessWithDetailedError(response, cancellationToken);

        var responseJson = await response.Content.ReadAsStringAsync(cancellationToken);
        return JsonSerializer.Deserialize<TResponse>(
            responseJson,
            AppJsonSerializerContext.Default.Options
        )!;
    }

    public async Task PostAsync<TRequest>(
        string endpoint,
        TRequest request,
        CancellationToken cancellationToken
    )
    {
        var json = JsonSerializer.Serialize(request, AppJsonSerializerContext.Default.Options);
        var content = new StringContent(json, Encoding.UTF8, "application/json");

        await AddAuthorizationHeader();

        var response = await httpClient.PostAsync(endpoint, content, cancellationToken);
        await EnsureSuccessWithDetailedError(response, cancellationToken);
    }

    public async Task<TResponse> GetAsync<TResponse>(
        string endpoint,
        CancellationToken cancellationToken
    )
    {
        await AddAuthorizationHeader();

        var response = await httpClient.GetAsync(endpoint, cancellationToken);
        await EnsureSuccessWithDetailedError(response, cancellationToken);

        var responseJson = await response.Content.ReadAsStringAsync(cancellationToken);
        return JsonSerializer.Deserialize<TResponse>(
            responseJson,
            AppJsonSerializerContext.Default.Options
        )!;
    }

    public async Task DeleteAsync(string endpoint, CancellationToken cancellationToken)
    {
        await AddAuthorizationHeader();

        var response = await httpClient.DeleteAsync(endpoint, cancellationToken);
        await EnsureSuccessWithDetailedError(response, cancellationToken);
    }

    private async Task AddAuthorizationHeader()
    {
        var userSession = await userSessionService.GetSession();
        if (userSession?.AccessToken != null)
        {
            if (userSession.Value.AccessTokenExpiresAt <= DateTime.UtcNow.AddMinutes(-15))
            {
                if (userSession.Value.RefreshTokenExpiresAt <= DateTime.UtcNow)
                {
                    navigationManager.NavigateTo("/login", forceLoad: true);
                    return;
                }

                var tokenDto = await PostAsync<RefreshRequestDto, TokenResponseDto>(
                    "/api/v1/auth/refresh",
                    new RefreshRequestDto { RefreshToken = userSession.Value.RefreshToken },
                    CancellationToken.None
                );

                userSession = new UserSession
                {
                    Username = userSession.Value.Username,
                    AccessToken = tokenDto.AccessToken,
                    AccessTokenExpiresAt = tokenDto.AccessTokenExpiresAt,
                    RefreshToken = tokenDto.RefreshToken,
                    RefreshTokenExpiresAt = tokenDto.RefreshTokenExpiresAt,
                };

                await authStateProvider.SetStateAsync(userSession.Value);
            }

            httpClient.DefaultRequestHeaders.Authorization =
                new System.Net.Http.Headers.AuthenticationHeaderValue(
                    "Bearer",
                    userSession.Value.AccessToken
                );
        }
    }

    private static async Task EnsureSuccessWithDetailedError(
        HttpResponseMessage response,
        CancellationToken cancellationToken
    )
    {
        if (response.IsSuccessStatusCode)
            return;

        string errorMessage;
        try
        {
            var errorJson = await response.Content.ReadAsStringAsync(cancellationToken);

            var errorResponse = JsonSerializer.Deserialize<ErrorDto>(
                errorJson,
                AppJsonSerializerContext.Default.Options
            );

            errorMessage = !string.IsNullOrEmpty(errorResponse?.Detail)
                ? errorResponse.Detail
                : response.ReasonPhrase ?? "An error occurred";
        }
        catch
        {
            errorMessage = response.ReasonPhrase ?? "An error occurred";
        }

        var exception = new HttpRequestException(errorMessage);
        exception.Data["StatusCode"] = response.StatusCode;
        throw exception;
    }
}
