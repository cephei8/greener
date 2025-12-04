using GreenerBlazor.Services;
using Microsoft.FluentUI.AspNetCore.Components;

namespace GreenerBlazor.Helpers;

public class ItemsProvider<TDto, TRequest, TResponse>(
    ExceptionService exceptionService,
    IItemsProviderAdapter<TDto, TRequest, TResponse> adapter
)
    where TRequest : class, new()
{
    public async ValueTask<GridItemsProviderResult<TDto>> HandleAsync(
        GridItemsProviderRequest<TDto> req
    )
    {
        var lookupRequest = new TRequest();
        adapter.SetOffset(lookupRequest, req.StartIndex);

        if (req.Count != null)
        {
            adapter.SetLimit(lookupRequest, req.Count.Value);
        }

        TResponse lookupResponse;
        try
        {
            lookupResponse = await adapter.MakeRequestAsync(lookupRequest, req.CancellationToken);
        }
        catch (HttpRequestException exc)
        {
            if (await exceptionService.HandleAsMessageBar(exc))
            {
                return GridItemsProviderResult.From<TDto>([], 0);
            }

            throw;
        }
        catch (OperationCanceledException)
        {
            return await ValueTask.FromCanceled<GridItemsProviderResult<TDto>>(
                req.CancellationToken
            );
        }

        await adapter.OnResponseFetched(lookupResponse);

        var rows = adapter.GetItems(lookupResponse);

        var totalCount = adapter.GetTotalCount(lookupResponse);

        return GridItemsProviderResult.From(rows, totalCount);
    }
}
