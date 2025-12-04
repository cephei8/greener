namespace GreenerBlazor.Helpers;

public interface IItemsProviderAdapter<TDto, in TRequest, TResponse>
    where TRequest : class
{
    public Func<TResponse, Task> OnResponseFetched { get; }

    string GetToken(TDto dto);

    void SetOffset(TRequest request, int offset);
    void SetLimit(TRequest request, int limit);
    Task<TResponse> MakeRequestAsync(TRequest request, CancellationToken cancellationToken);
    List<TDto> GetItems(TResponse response);
    int GetTotalCount(TResponse response);
}
