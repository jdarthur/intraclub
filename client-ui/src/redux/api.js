import {createApi, fetchBaseQuery} from '@reduxjs/toolkit/query/react';
import {logoutUser} from "./auth.js";

export const baseQuery = fetchBaseQuery({
    baseUrl: '/api/',
    prepareHeaders: (headers, {getState}) => {
        const userToken = getState().auth.token
        if (userToken) {
            headers.set('X-INTRACLUB-TOKEN', userToken);
        }

        return headers;
    }
});

const baseQueryWithLogout = async (args, api, extraOptions) => {
    let result = await baseQuery(args, api, extraOptions)
    if (result?.error && result?.error?.status === 401)
    {
        api.dispatch(logoutUser());
    }
    return result
}


export const mainApi = createApi({
    reducerPath: 'mainApi',
    baseQuery: baseQueryWithLogout,
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
            query: () => ({url: `user`}),
            providesTags: ['users']
        }),
        getUserById: builder.query({
            query: (userId) => ({url: `user/${userId}`}),
            providesTags: ["user_by_id"]
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
            query: () => ({url: `facility`}),
            providesTags: ["facility"]
        }),
        createFacility: builder.mutation({
            query: (body) => ({url: `facility`, body: body, method: 'POST'}),
            invalidatesTags: ["facility"]
        }),
        updateFacility: builder.mutation({
            query: (req) => ({url: `facility/${req.id}`, body: req.body, method: 'put'}),
            invalidatesTags: ["facility"]
        }),
        deleteFacility: builder.mutation({
            query: (id) => ({url: `facility/${id}`, method: 'DELETE'}),
            invalidatesTags: ["facility"]
        }),
        createWeek: builder.mutation({
            query: (body) => ({url: `week`, method: 'POST', body: body}),
            invalidatesTags: ["weeks", "league_weeks"]
        }),
        createLeague: builder.mutation({
            query: (body) => ({url: `league`, method: 'POST', body: body}),
            invalidatesTags: ["leagues"]
        }),
        updateLeague: builder.mutation({
            query: (args) => ({url: `league/${args.id}`, method: 'PUT', body: args.body}),
            invalidatesTags: ["leagues", "league"]
        }),
        getWeekById: builder.query({
            query: (id) => ({url: `weeks/${id}`}),
        }),
        getWeeksByLeagueId: builder.query({
            query: (id) => ({url: `league/${id}/weeks`}),
            providesTags: ["league_weeks"]
        }),
        deleteWeek: builder.mutation({
            query: (id) => ({url: `week/${id}`, method: 'DELETE'}),
            invalidatesTags: ["weeks", "league_weeks", "league"]
        }),
        getFacilityById: builder.query({
            query: (id) => ({url: `facility/${id}`}),
        }),
        deleteLeague: builder.mutation({
            query: (id) => ({url: `league/${id}`, method: 'DELETE'}),
            invalidatesTags: ["leagues", "league"]
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
        getSkillInfo: builder.query({
            query: (id) => ({url: `skill_info_for_user/${id}`}),
            providesTags: ["skill_info"]
        }),
        getSkillInfoOptions: builder.query({
            query: () => ({url: `skill_info_options`}),
            providesTags: ["skill_info_options"]
        }),
        createSkillInfo: builder.mutation({
            query: (body) => ({url: `skill_info`, method: "POST", body: body}),
            invalidatesTags: ["skill_info"]
        }),
        deleteSkillInfo: builder.mutation({
            query: (id) => ({url: `skill_info/${id}`, method: "DELETE"}),
            invalidatesTags: ["skill_info"]
        }),
        getLeague: builder.query({
            query: (id) => ({url: `league/${id}`}),
            providesTags: ["league"]
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
    useGetSkillInfoQuery,
    useCreateSkillInfoMutation,
    useGetSkillInfoOptionsQuery,
    useDeleteSkillInfoMutation,
    useGetLeagueQuery,
} = mainApi
