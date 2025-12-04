using GreenerBlazor.Models;

namespace GreenerBlazor.Services;

public class ApiKeyService(ApiClient apiClient)
{
    public async Task<PaginatedResponseDto<ApiKeyDto>> GetApiKeysAsync(
        int offset,
        int limit,
        CancellationToken cancellationToken
    )
    {
        var endpoint = $"/api/v1/api-keys?offset={offset}&limit={limit}";
        return await apiClient.GetAsync<PaginatedResponseDto<ApiKeyDto>>(
            endpoint,
            cancellationToken
        );
    }

    public async Task<CreateApiKeyResponseDto> CreateApiKeyAsync(
        CreateApiKeyRequestDto request,
        CancellationToken cancellationToken
    )
    {
        return await apiClient.PostAsync<CreateApiKeyRequestDto, CreateApiKeyResponseDto>(
            "/api/v1/api-keys",
            request,
            cancellationToken
        );
    }

    public async Task<ApiKeyDto> GetApiKeyAsync(string uid, CancellationToken cancellationToken)
    {
        return await apiClient.GetAsync<ApiKeyDto>($"/api/v1/api-keys/{uid}", cancellationToken);
    }

    public async Task DeleteApiKeyAsync(string uid, CancellationToken cancellationToken)
    {
        await apiClient.DeleteAsync($"/api/v1/api-keys/{uid}", cancellationToken);
    }
}
