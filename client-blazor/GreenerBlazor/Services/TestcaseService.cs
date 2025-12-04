using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class TestcaseService(ApiClient apiClient)
{
    public async Task<TestcasePaginatedResponseDto> GetTestcasesAsync(
        int offset,
        int limit,
        string? query,
        DateTime? startDate,
        DateTime? endDate,
        string? group,
        CancellationToken cancellationToken
    )
    {
        var queryParams = new List<string> { $"offset={offset}", $"limit={limit}" };

        if (!string.IsNullOrEmpty(query))
            queryParams.Add($"queryStr={Uri.EscapeDataString(query)}");

        if (startDate.HasValue)
            queryParams.Add($"startDate={startDate.Value:yyyy-MM-ddTHH:mm:ssZ}");

        if (endDate.HasValue)
            queryParams.Add($"endDate={endDate.Value:yyyy-MM-ddTHH:mm:ssZ}");

        if (!string.IsNullOrEmpty(group))
            queryParams.Add($"group={Uri.EscapeDataString(group)}");

        var endpoint = $"/api/v1/testcases?{string.Join("&", queryParams)}";
        return await apiClient.GetAsync<TestcasePaginatedResponseDto>(
            endpoint,
            cancellationToken
        );
    }

    public async Task<TestcaseDto> GetTestcaseAsync(string uid, CancellationToken cancellationToken)
    {
        return await apiClient.GetAsync<TestcaseDto>($"/api/v1/testcases/{uid}", cancellationToken);
    }
}
