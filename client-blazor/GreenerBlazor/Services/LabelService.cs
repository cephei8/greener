using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class LabelService(ApiClient apiClient)
{
    public async Task<PaginatedResponseDto<LabelDto>> GetLabelsBySessionAsync(
        string sessionId,
        int offset,
        int limit,
        CancellationToken cancellationToken
    )
    {
        var endpoint = $"/api/v1/labels?session_id={sessionId}&offset={offset}&limit={limit}";
        return await apiClient.GetAsync<PaginatedResponseDto<LabelDto>>(
            endpoint,
            cancellationToken
        );
    }
}
