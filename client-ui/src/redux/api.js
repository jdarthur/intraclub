import {createApi, fetchBaseQuery} from '@reduxjs/toolkit/query/react';

export const baseQuery = fetchBaseQuery({
    baseUrl: '/api/',
    prepareHeaders: (headers, {getState}) => {
        const userToken = getState().auth.token
        if (userToken) {
            headers.set('x-session-token', userToken);
        }

        return headers;
    }
});

export const mainApi = createApi({
    reducerPath: 'mainApi',
    baseQuery: baseQuery,
    endpoints: (builder) => ({
        createOneTimePassword: builder.mutation({
            query: (body) => ({url: `one_time_password`, method: 'POST', body: body}),
        }),
        register: builder.mutation({
            query: (body) => ({url: `register`, method: 'POST', body: body}),
        }),
        getToken: builder.mutation({
            query: (body) => ({
                url: 'token',
                method: 'POST',
                body: body,
            })
        }),
        whoAmI: builder.query({
            query: () => ({url: 'whoami'}),
            providesTags: ['whoami']
        }),
        getUsers: builder.query({
            query: () => ({url: `users`}),
            providesTags: ['users']
        }),
        getUserById: builder.query({
            query: (userId) => ({url: `users/${userId}`}),
        }),
        getTeamsByUserId: builder.query({
            query: (userId) => ({url: `teams_for_user/${userId}`}),
        }),
        getLeaguesByUserId: builder.query({
            query: (userId) => ({url: `leagues_for_user/${userId}`}),
            providesTags: ["leagues"]
        }),
        getLeaguesCommissionedByUserId: builder.query({
            query: (userId) => ({url: `leagues_commissioned_by_user/${userId}`}),
            providesTags: ["leagues"]
        }),
        getFacilities: builder.query({
            query: () => ({url: `facilities`}),
            providesTags: ["facilities"]
        }),
        createFacility: builder.mutation({
            query: (body) => ({url: `facilities`, body: body, method: 'POST'}),
            invalidatesTags: ["facilities"]
        }),
        updateFacility: builder.mutation({
            query: (req) => ({url: `facilities/${req.id}`, body: req.body, method: 'put'}),
            invalidatesTags: ["facilities"]
        }),
        deleteFacility: builder.mutation({
            query: (id) => ({url: `facilities/${id}`, method: 'DELETE'}),
            invalidatesTags: ["facilities"]
        }),
        createWeek: builder.mutation({
            query: (body) => ({url: `weeks`, method: 'POST', body: body}),
            invalidatesTags: ["weeks", "league_weeks"]
        }),
        createLeague: builder.mutation({
            query: (body) => ({url: `leagues`, method: 'POST', body: body}),
            invalidatesTags: ["leagues"]
        }),
        updateLeague: builder.mutation({
            query: (args) => ({url: `leagues/${args.id}`, method: 'PUT', body: args.body}),
            invalidatesTags: ["leagues"]
        }),
        getWeekById: builder.query({
            query: (id) => ({url: `weeks/${id}`}),
        }),
        getWeeksByLeagueId: builder.query({
            query: (id) => ({url: `league/${id}/weeks`}),
            providesTags: ["league_weeks"]
        }),
        deleteWeek: builder.mutation({
            query: (id) => ({url: `weeks/${id}`, method: 'DELETE'}),
            invalidatesTags: ["weeks", "league_weeks"]
        }),
        getFacilityById: builder.query({
            query: (id) => ({url: `facilities/${id}`}),
        }),
        deleteLeague: builder.mutation({
            query: (id) => ({url: `leagues/${id}`, method: 'DELETE'}),
            invalidatesTags: ["leagues"]
        }),
        // get multiple weeks by a list of IDs
        getWeeksByIds: builder.query({
            query: (weekIds) => ({url: `weeks_search`, method: "POST", body: weekIds}),
            providesTags: ["league_weeks"]
        }),
        importUsers: builder.mutation({
            query: (body) => {
                const f = new FormData()
                f.append("file", body)
                return {
                    url: `import_users_from_csv`,
                    method: "POST",
                    body: f,
                    formData: true
                }
            },
            invalidatesTags: ["users"]
        }),
        getMatchScores: builder.query({
            query: () => ({url: `match_scores`}),
            providesTags: ["match_scores"]
        }),
        updateMatchScoresForLine: builder.mutation({
            query: (body) => ({url: `match_scores?key=${body.key}`, method: "PUT", body: body}),
            invalidatesTags: ["match_scores"]
        }),
        updateNameForLine: builder.mutation({
            query: (body) => ({url: `match_player_names?key=${body.key}`, method: "PUT", body: body}),
            invalidatesTags: ["match_scores"]
        }),
        updateTeamInfo: builder.mutation({
            query: (body) => ({url: `match_team_info?key=${body.key}`, method: "PUT", body: body}),
            invalidatesTags: ["match_scores"]
        }),
    })
});

export const {
    useGetUsersQuery,
    useCreateOneTimePasswordMutation,
    useGetTokenMutation,
    useRegisterMutation,
    useWhoAmIQuery,
    useGetUserByIdQuery,
    useGetTeamsByUserIdQuery,
    useGetLeaguesByUserIdQuery,
    useGetLeaguesCommissionedByUserIdQuery,
    useGetFacilitiesQuery,
    useCreateFacilityMutation,
    useDeleteFacilityMutation,
    useUpdateFacilityMutation,
    useCreateWeekMutation,
    useCreateLeagueMutation,
    useUpdateLeagueMutation,
    useGetWeekByIdQuery,
    useGetFacilityByIdQuery,
    useDeleteLeagueMutation,
    useDeleteWeekMutation,
    useGetWeeksByIdsQuery,
    useImportUsersMutation,
    useLazyGetMatchScoresQuery,
    useUpdateMatchScoresForLineMutation,
    useUpdateNameForLineMutation,
    useUpdateTeamInfoMutation,
} = mainApi
