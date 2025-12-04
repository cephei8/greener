using Blazored.LocalStorage;
using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class UserSessionService(ILocalStorageService localStorage)
{
    private const string SessionKey = "user-session";

    public async Task<UserSession?> GetSession()
    {
        if (!await localStorage.ContainKeyAsync(SessionKey))
        {
            return null;
        }

        return await localStorage.GetItemAsync<UserSession>(SessionKey);
    }

    public async Task SetSession(UserSession session) =>
        await localStorage.SetItemAsync(SessionKey, session);

    public async Task RemoveSession() => await localStorage.RemoveItemAsync(SessionKey);
}
