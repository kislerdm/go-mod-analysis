package parsehtml_test

import (
	_ "embed"
	"gomodanalysis/cmd/listmodules/parsehtml"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestParseHtml(t *testing.T) {
	type args struct {
		v io.Reader
	}

	tests := []struct {
		name    string
		args    args
		wantO   []parsehtml.HTMLCodeSearchContent
		wantErr bool
	}{
		{
			name: "happy path: array of 1",
			args: args{
				v: strings.NewReader(`<div 
  class="hx_hit-code code-list-item d-flex py-4 code-list-item-public ">
    <img class="rounded-2 v-align-middle flex-shrink-0 mr-1" src="https://avatars.githubusercontent.com/u/50180778?s=40&amp;v=4" width="20" height="20" alt="@trustbloc" />

  <div class="width-full">
      <div class="flex-shrink-0 text-small text-bold">
        <a class="Link--secondary" href="/trustbloc/edv">
          trustbloc/edv
</a>      </div>

    <div class="f4 text-normal">
      <a title="/go.mod" data-hydro-click="{&quot;event_type&quot;:&quot;search_result.click&quot;,&quot;payload&quot;:{&quot;page_number&quot;:11,&quot;per_page&quot;:10,&quot;query&quot;:&quot;github.com/google/tink/go filename:go.mod&quot;,&quot;result_position&quot;:1,&quot;click_id&quot;:232844851,&quot;result&quot;:{&quot;id&quot;:232844851,&quot;global_relay_id&quot;:&quot;MDEwOlJlcG9zaXRvcnkyMzI4NDQ4NTE=&quot;,&quot;model_name&quot;:&quot;Repository&quot;,&quot;url&quot;:&quot;https://github.com/trustbloc/edv/blob/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod&quot;},&quot;originating_url&quot;:&quot;https://github.com/search?q=github.com%2Fgoogle%2Ftink%2Fgo+filename%3Ago.mod&amp;p=11&quot;,&quot;user_id&quot;:13434797}}" data-hydro-click-hmac="9756aea896d66a43ed92ef28537366848408374654e3b5dd1f8e4e9731782c29" href="/trustbloc/edv/blob/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod">/<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go.mod</span></a>
    </div>

      <div class="file-box blob-wrapper my-1">
        <table class="highlight">
            <tr>
    <td class="blob-num">
      <a href="/trustbloc/edv/blob/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod#L44">44</a>
    </td>
    <td class="blob-code blob-code-inner">	github.com/google/go-cmp v0.5.6 // indirect
</td>
  </tr>
  <tr>
    <td class="blob-num">
      <a href="/trustbloc/edv/blob/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod#L45">45</a>
    </td>
    <td class="blob-code blob-code-inner">	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
</td>
  </tr>
  <tr>
    <td class="blob-num">
      <a href="/trustbloc/edv/blob/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod#L46">46</a>
    </td>
    <td class="blob-code blob-code-inner">	<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>github</span>.<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>com</span>/<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>google</span>/<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>tink</span>/<span class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go</span> v1.6.1-0.20210519071714-58be99b3c4d0 // indirect</td>
  </tr>

        </table>
      </div>

    <div class="d-flex flex-wrap text-small color-fg-muted">
         <div class="mr-3">
           <span class="">
  <span class="repo-language-color" style="background-color: #ccc"></span>
  <span itemprop="programmingLanguage">Text</span>
</span>

         </div>

        <span class="match-count mr-3">
          Showing the top five matches
        </span>

      <span class="updated-at mr-3">
        Last indexed <relative-time datetime="2022-07-28T08:10:47Z" class="no-wrap">Jul 28, 2022</relative-time>
      </span>
    </div>

  </div>
</div>`),
			},
			wantO: []parsehtml.HTMLCodeSearchContent{
				{
					Repo:     "/trustbloc/edv",
					FilePath: "/trustbloc/edv/eca059d01e5049f58997f59afbd3a74f572df1de/go.mod",
				},
			},
			wantErr: false,
		},
		{
			name: "happy path: array of 3",
			args: args{
				v: strings.NewReader(`<div
                                    class="hx_hit-code code-list-item d-flex py-4 code-list-item-public ">
                                <img class="rounded-2 v-align-middle flex-shrink-0 mr-1"
                                     src="https://avatars.githubusercontent.com/u/50180778?s=40&amp;v=4" width="20"
                                     height="20" alt="@trustbloc"/>

                                <div class="width-full">
                                    <div class="flex-shrink-0 text-small text-bold">
                                        <a class="Link--secondary" href="/trustbloc/agent-sdk">
                                            trustbloc/agent-sdk
                                        </a></div>

                                    <div class="f4 text-normal">
                                        <a title="cmd/agent-mobile/go.mod"
                                           data-hydro-click="{&quot;event_type&quot;:&quot;search_result.click&quot;,&quot;payload&quot;:{&quot;page_number&quot;:12,&quot;per_page&quot;:10,&quot;query&quot;:&quot;github.com/google/tink/go filename:go.mod&quot;,&quot;result_position&quot;:1,&quot;click_id&quot;:302138629,&quot;result&quot;:{&quot;id&quot;:302138629,&quot;global_relay_id&quot;:&quot;MDEwOlJlcG9zaXRvcnkzMDIxMzg2Mjk=&quot;,&quot;model_name&quot;:&quot;Repository&quot;,&quot;url&quot;:&quot;https://github.com/trustbloc/agent-sdk/blob/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod&quot;},&quot;originating_url&quot;:&quot;https://github.com/search?q=github.com%2Fgoogle%2Ftink%2Fgo+filename%3Ago.mod&amp;p=12&quot;,&quot;user_id&quot;:13434797}}"
                                           data-hydro-click-hmac="fcf3d315989a428f5ba6f8fcfe95268b46e94c12bf1f51b65ed5b84e0da30af4"
                                           href="/trustbloc/agent-sdk/blob/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod">cmd/agent-mobile/<span
                                                class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go.mod</span></a>
                                    </div>

                                    <div class="file-box blob-wrapper my-1">
                                        <table class="highlight">
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/agent-sdk/blob/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod#L41">41</a>
                                                </td>
                                                <td class="blob-code blob-code-inner">
                                                    github.com/google/certificate-transparency-go
                                                    v1.1.2-0.20210512142713-bed466244fa6 // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/agent-sdk/blob/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod#L42">42</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"><span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>github</span>.<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>com</span>/google/tink/go
                                                    v1.6.1 // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/agent-sdk/blob/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod#L43">43</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"> github.com/google/trillian
                                                    v1.3.14-0.20210520152752-ceda464a95a3 // indirect
                                                </td>
                                            </tr>

                                        </table>
                                    </div>

                                    <div class="d-flex flex-wrap text-small color-fg-muted">
                                        <div class="mr-3">
           <span class="">
  <span class="repo-language-color" style="background-color: #ccc"></span>
  <span itemprop="programmingLanguage">Text</span>
</span>

                                        </div>

                                        <span class="match-count mr-3">
          Showing the top two matches
        </span>

                                        <span class="updated-at mr-3">
        Last indexed <relative-time datetime="2022-10-05T13:18:04Z" class="no-wrap">Oct 5, 2022</relative-time>
      </span>
                                    </div>

                                </div>
                            </div>


                            <div
                                    class="hx_hit-code code-list-item d-flex py-4 code-list-item-public ">
                                <img class="rounded-2 v-align-middle flex-shrink-0 mr-1"
                                     src="https://avatars.githubusercontent.com/u/50180778?s=40&amp;v=4" width="20"
                                     height="20" alt="@trustbloc"/>

                                <div class="width-full">
                                    <div class="flex-shrink-0 text-small text-bold">
                                        <a class="Link--secondary" href="/trustbloc/adapter">
                                            trustbloc/adapter
                                        </a></div>

                                    <div class="f4 text-normal">
                                        <a title="/go.mod"
                                           data-hydro-click="{&quot;event_type&quot;:&quot;search_result.click&quot;,&quot;payload&quot;:{&quot;page_number&quot;:12,&quot;per_page&quot;:10,&quot;query&quot;:&quot;github.com/google/tink/go filename:go.mod&quot;,&quot;result_position&quot;:2,&quot;click_id&quot;:235842673,&quot;result&quot;:{&quot;id&quot;:235842673,&quot;global_relay_id&quot;:&quot;MDEwOlJlcG9zaXRvcnkyMzU4NDI2NzM=&quot;,&quot;model_name&quot;:&quot;Repository&quot;,&quot;url&quot;:&quot;https://github.com/trustbloc/adapter/blob/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod&quot;},&quot;originating_url&quot;:&quot;https://github.com/search?q=github.com%2Fgoogle%2Ftink%2Fgo+filename%3Ago.mod&amp;p=12&quot;,&quot;user_id&quot;:13434797}}"
                                           data-hydro-click-hmac="62aae0d1fbcb67db9434da4469c0d4761b5e262018f2e40db2fbe5a43e3a5ede"
                                           href="/trustbloc/adapter/blob/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod">/<span
                                                class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go.mod</span></a>
                                    </div>

                                    <div class="file-box blob-wrapper my-1">
                                        <table class="highlight">
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/adapter/blob/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod#L57">57</a>
                                                </td>
                                                <td class="blob-code blob-code-inner">
                                                    github.com/google/certificate-transparency-go
                                                    v1.1.2-0.20210512142713-bed466244fa6 // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/adapter/blob/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod#L58">58</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"><span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>github</span>.<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>com</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>google</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>tink</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go</span>
                                                    v1.6.1 // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/trustbloc/adapter/blob/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod#L59">59</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"> github.com/google/trillian
                                                    v1.3.14-0.20210520152752-ceda464a95a3 // indirect
                                                </td>
                                            </tr>

                                        </table>
                                    </div>

                                    <div class="d-flex flex-wrap text-small color-fg-muted">
                                        <div class="mr-3">
           <span class="">
  <span class="repo-language-color" style="background-color: #ccc"></span>
  <span itemprop="programmingLanguage">Text</span>
</span>

                                        </div>

                                        <span class="match-count mr-3">
          Showing the top five matches
        </span>

                                        <span class="updated-at mr-3">
        Last indexed <relative-time datetime="2022-08-31T14:40:46Z" class="no-wrap">Aug 31, 2022</relative-time>
      </span>
                                    </div>

                                </div>
                            </div>


                            <div
                                    class="hx_hit-code code-list-item d-flex py-4 code-list-item-public ">
                                <img class="rounded-2 v-align-middle flex-shrink-0 mr-1"
                                     src="https://avatars.githubusercontent.com/u/102382658?s=40&amp;v=4" width="20"
                                     height="20" alt="@ice-blockchain"/>

                                <div class="width-full">
                                    <div class="flex-shrink-0 text-small text-bold">
                                        <a class="Link--secondary" href="/ice-blockchain/wintr">
                                            ice-blockchain/wintr
                                        </a></div>

                                    <div class="f4 text-normal">
                                        <a title="/go.mod"
                                           data-hydro-click="{&quot;event_type&quot;:&quot;search_result.click&quot;,&quot;payload&quot;:{&quot;page_number&quot;:12,&quot;per_page&quot;:10,&quot;query&quot;:&quot;github.com/google/tink/go filename:go.mod&quot;,&quot;result_position&quot;:3,&quot;click_id&quot;:474780797,&quot;result&quot;:{&quot;id&quot;:474780797,&quot;global_relay_id&quot;:&quot;R_kgDOHEyUfQ&quot;,&quot;model_name&quot;:&quot;Repository&quot;,&quot;url&quot;:&quot;https://github.com/ice-blockchain/wintr/blob/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod&quot;},&quot;originating_url&quot;:&quot;https://github.com/search?q=github.com%2Fgoogle%2Ftink%2Fgo+filename%3Ago.mod&amp;p=12&quot;,&quot;user_id&quot;:13434797}}"
                                           data-hydro-click-hmac="47af5255e7000ca1695ff599a98dfbb0358ac0844ddc8005996a832096ecdb01"
                                           href="/ice-blockchain/wintr/blob/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod">/<span
                                                class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go.mod</span></a>
                                    </div>

                                    <div class="file-box blob-wrapper my-1">
                                        <table class="highlight">
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/ice-blockchain/wintr/blob/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod#L71">71</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"> github.com/golang/protobuf v1.5.2
                                                    // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/ice-blockchain/wintr/blob/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod#L72">72</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"> github.com/google/go-cmp v0.5.9
                                                    // indirect
                                                </td>
                                            </tr>
                                            <tr>
                                                <td class="blob-num">
                                                    <a href="/ice-blockchain/wintr/blob/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod#L73">73</a>
                                                </td>
                                                <td class="blob-code blob-code-inner"><span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>github</span>.<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>com</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>google</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>tink</span>/<span
                                                        class='text-bold hx_keyword-hl rounded-2 d-inline-block'>go</span>
                                                    v1.7.0 // indirect
                                                </td>
                                            </tr>

                                        </table>
                                    </div>

                                    <div class="d-flex flex-wrap text-small color-fg-muted">
                                        <div class="mr-3">
           <span class="">
  <span class="repo-language-color" style="background-color: #ccc"></span>
  <span itemprop="programmingLanguage">Text</span>
</span>

                                        </div>

                                        <span class="match-count mr-3">
          Showing the top five matches
        </span>

                                        <span class="updated-at mr-3">
        Last indexed <relative-time datetime="2022-10-05T00:24:05Z" class="no-wrap">Oct 5, 2022</relative-time>
      </span>
                                    </div>

                                </div>
                            </div>`),
			},
			wantO: []parsehtml.HTMLCodeSearchContent{
				{
					Repo:     "/trustbloc/agent-sdk",
					FilePath: "/trustbloc/agent-sdk/a66e47e80ee94744786adc153defb636a0aac8ee/cmd/agent-mobile/go.mod",
				},
				{
					Repo:     "/trustbloc/adapter",
					FilePath: "/trustbloc/adapter/3d5563678a0ccc4a6bf8dbc031c3e526a4bdbcdb/go.mod",
				},
				{
					Repo:     "/ice-blockchain/wintr",
					FilePath: "/ice-blockchain/wintr/e634068ae9b088c52ccd2eafa9b068934a558cd0/go.mod",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotO, err := parsehtml.ParseHtml(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHtml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotO, tt.wantO) {
				t.Errorf("ParseHtml() gotO = %v, want %v", gotO, tt.wantO)
			}
		})
	}
}
