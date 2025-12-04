using System.Text.Json;
using System.Text.Json.Serialization;
using GreenerBlazor.Models;

namespace GreenerBlazor
{
    [JsonSourceGenerationOptions(
        RespectNullableAnnotations = true,
        PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
        DictionaryKeyPolicy = JsonKnownNamingPolicy.CamelCase
    )]
    [JsonSerializable(typeof(JsonDocument))]
    [JsonSerializable(typeof(ErrorDto))]
    [JsonSerializable(typeof(LoginRequestDto))]
    [JsonSerializable(typeof(TokenResponseDto))]
    [JsonSerializable(typeof(ChangePasswordRequestDto))]
    [JsonSerializable(typeof(TestcaseStatus))]
    [JsonSerializable(typeof(TestcaseDto))]
    [JsonSerializable(typeof(SessionDto))]
    [JsonSerializable(typeof(ApiKeyDto))]
    [JsonSerializable(typeof(CreateApiKeyRequestDto))]
    [JsonSerializable(typeof(CreateApiKeyResponseDto))]
    [JsonSerializable(typeof(QueryValidationResult))]
    [JsonSerializable(typeof(TestcasePaginatedResponseDto))]
    [JsonSerializable(typeof(LabelDto))]
    [JsonSerializable(typeof(PaginatedResponseDto<LabelDto>))]
    [JsonSerializable(typeof(PaginatedResponseDto<SessionDto>))]
    [JsonSerializable(typeof(PaginatedResponseDto<ApiKeyDto>))]
    [JsonSerializable(typeof(GroupItemDto))]
    [JsonSerializable(typeof(GroupPaginatedResponseDto))]
    internal partial class AppJsonSerializerContext : JsonSerializerContext;
}
