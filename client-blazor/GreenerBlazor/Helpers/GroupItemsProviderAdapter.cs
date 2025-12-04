using GreenerBlazor.Models;
using GreenerBlazor.Services;

namespace GreenerBlazor.Helpers;

public class GroupItemsProviderAdapter(
    GroupService groupApiService,
    Func<GroupPaginatedResponseDto, Task> onResponseFetched,
    string? query,
    DateTime? startDate,
    DateTime? endDate
) : IItemsProviderAdapter<GroupRow, GroupListRequest, GroupPaginatedResponseDto>
{
    public Func<GroupPaginatedResponseDto, Task> OnResponseFetched => onResponseFetched;

    public string GetToken(GroupRow row) => row.Id;

    public void SetOffset(GroupListRequest request, int offset) => request.Offset = offset;

    public void SetLimit(GroupListRequest request, int limit) => request.Limit = limit;

    public async Task<GroupPaginatedResponseDto> MakeRequestAsync(
        GroupListRequest request,
        CancellationToken cancellationToken
    )
    {
        if (string.IsNullOrEmpty(query))
        {
            return new GroupPaginatedResponseDto
            {
                Items = [],
                Total = 0,
                Limit = request.Limit,
                Offset = request.Offset,
                Header = null,
                AggregatedStatus = null,
            };
        }

        return await groupApiService.GetGroupsAsync(
            query,
            request.Offset,
            request.Limit,
            startDate,
            endDate,
            cancellationToken
        );
    }

    public List<GroupRow> GetItems(GroupPaginatedResponseDto response)
    {
        return
        [
            .. response.Items.Select(g =>
            {
                return new GroupRow(
                    string.Join(" | ", g.Columns.Select(c => c ?? "(none)")),
                    g.Status.ToString(),
                    response.Header!,
                    g.Columns
                );
            }),
        ];
    }

    public int GetTotalCount(GroupPaginatedResponseDto response) => response.Total;
}

public class GroupListRequest
{
    public int Offset { get; set; }
    public int Limit { get; set; }
}
