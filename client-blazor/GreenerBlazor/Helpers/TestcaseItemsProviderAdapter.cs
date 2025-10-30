using GreenerBlazor.Models;
using GreenerBlazor.Services;

namespace GreenerBlazor.Helpers;

public class TestcaseRestItemsProviderAdapter(
    TestcaseService testcaseApiService,
    Func<TestcasePaginatedResponseDto, Task> onResponseFetched,
    string? query,
    DateTime? startDate,
    DateTime? endDate,
    string? group
) : IItemsProviderAdapter<TestcaseRow, TestcaseListRequest, TestcasePaginatedResponseDto>
{
    public Func<TestcasePaginatedResponseDto, Task> OnResponseFetched => onResponseFetched;

    public string GetToken(TestcaseRow row) => row.Id;

    public void SetOffset(TestcaseListRequest request, int offset) => request.Offset = offset;

    public void SetLimit(TestcaseListRequest request, int limit) => request.Limit = limit;

    public async Task<TestcasePaginatedResponseDto> MakeRequestAsync(
        TestcaseListRequest request,
        CancellationToken cancellationToken
    )
    {
        return await testcaseApiService.GetTestcasesAsync(
            request.Offset,
            request.Limit,
            query,
            startDate,
            endDate,
            group,
            cancellationToken
        );
    }

    public List<TestcaseRow> GetItems(TestcasePaginatedResponseDto response) =>
        [
            .. response.Items.Select(tc => new TestcaseRow(
                tc.Id,
                tc.SessionId.ToString(),
                tc.Name,
                tc.Status
            )),
        ];

    public int GetTotalCount(TestcasePaginatedResponseDto response) => response.Total;
}

public class TestcaseListRequest
{
    public int Offset { get; set; }
    public int Limit { get; set; }
}
