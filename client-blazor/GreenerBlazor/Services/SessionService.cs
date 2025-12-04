using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class SessionService(ApiClient apiClient)
{
    public async Task<PaginatedResponseDto<SessionDto>> GetSessionsAsync(
        int offset,
        int limit,
        CancellationToken cancellationToken
    )
    {
        var endpoint = $"/api/v1/sessions?offset={offset}&limit={limit}";
        return await apiClient.GetAsync<PaginatedResponseDto<SessionDto>>(
            endpoint,
            cancellationToken
        );
    }

    public async Task<SessionDto> GetSessionAsync(string uid, CancellationToken cancellationToken)
    {
        return await apiClient.GetAsync<SessionDto>($"/api/v1/sessions/{uid}", cancellationToken);
    }
}
