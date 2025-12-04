using System.Security.Claims;
using GreenerBlazor.Models;
using Microsoft.AspNetCore.Components.Authorization;

namespace GreenerBlazor.Services;

public class AppAuthenticationStateProvider(UserSessionService userSessionService)
    : AuthenticationStateProvider
{
    public override async Task<AuthenticationState> GetAuthenticationStateAsync()
    {
        var userSession = await userSessionService.GetSession();
        return await CreateAuthenticationStateAsync(userSession);
    }

    public async Task<UserSession?> GetUserSessionAsync() => await userSessionService.GetSession();

    public async Task SetStateAsync(UserSession userSession)
    {
        await userSessionService.SetSession(userSession);
        NotifyAuthenticationStateChanged(CreateAuthenticationStateAsync(userSession));
    }

    public async Task RemoveStateAsync()
    {
        await userSessionService.RemoveSession();
        NotifyAuthenticationStateChanged(CreateAuthenticationStateAsync());
    }

    private static async Task<AuthenticationState> CreateAuthenticationStateAsync(
        UserSession? userSession = null
    )
    {
        if (userSession == null)
        {
            return await Task.FromResult(
                new AuthenticationState(new ClaimsPrincipal(new ClaimsIdentity()))
            );
        }

        var claimsPrincipal = new ClaimsPrincipal(
            new ClaimsIdentity([new(ClaimTypes.Name, userSession.Value.Username)], "Bearer")
        );

        return await Task.FromResult(new AuthenticationState(claimsPrincipal));
    }
}
