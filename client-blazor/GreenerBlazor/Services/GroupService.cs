using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class GroupService(ApiClient apiClient)
{
    public async Task<QueryValidationResult> ValidateQueryAsync(
        string query,
        CancellationToken cancellationToken
    )
    {
        var endpoint = $"/api/v1/groups/validate-query?queryStr={Uri.EscapeDataString(query)}";
        return await apiClient.GetAsync<QueryValidationResult>(endpoint, cancellationToken);
    }

    public async Task<GroupPaginatedResponseDto> GetGroupsAsync(
        string query,
        int offset,
        int limit,
        DateTime? startDate,
        DateTime? endDate,
        CancellationToken cancellationToken
    )
    {
        var queryParams = new List<string>
        {
            $"queryStr={Uri.EscapeDataString(query)}",
            $"offset={offset}",
            $"limit={limit}",
        };

        if (startDate.HasValue)
            queryParams.Add($"startDate={startDate.Value:yyyy-MM-ddTHH:mm:ssZ}");

        if (endDate.HasValue)
            queryParams.Add($"endDate={endDate.Value:yyyy-MM-ddTHH:mm:ssZ}");

        var endpoint = $"/api/v1/groups?{string.Join("&", queryParams)}";
        return await apiClient.GetAsync<GroupPaginatedResponseDto>(endpoint, cancellationToken);
    }
}
