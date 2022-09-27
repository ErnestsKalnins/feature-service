import {Injectable} from '@angular/core';
import {HttpClient, HttpResponse} from "@angular/common/http";
import {Observable} from "rxjs";
import {Feature} from "./feature";
import {environment} from "../../../environments/environment.prod";

@Injectable({
  providedIn: 'root'
})
export class FeatureService {
  private featuresUrl = `${environment.apiHost}/api/v1/features`;
  private archivedFeaturesUrl = `${environment.apiHost}/api/v1/archived_features`;

  constructor(
    private http: HttpClient,
  ) {
  }

  getFeatures(): Observable<Feature[]> {
    return this.http.get<Feature[]>(this.featuresUrl);
  }

  getFeature(id: string): Observable<Feature> {
    return this.http.get<Feature>(this.featuresUrl + `/${id}`);
  }

  saveFeature({
                technicalName,
                displayName,
                description,
                inverted,
                expiresOn,
                customerIds
              }: Feature): Observable<HttpResponse<void>> {
    const expiresOnRFC3339 = expiresOn === null
      ? null
      : new Date(expiresOn);

    return this.http.post<HttpResponse<void>>(this.featuresUrl, {
      technicalName,
      displayName,
      description,
      inverted,
      customerIds,
      expiresOn: expiresOn === null ? undefined : new Date(expiresOn).valueOf()
    });
  }

  updateFeature({
                  id,
                  updatedAt,
                  displayName,
                  technicalName,
                  expiresOn,
                  description,
                  inverted,
                  customerIds,
                }: Feature): Observable<HttpResponse<void>> {
    return this.http.put<HttpResponse<void>>(this.featuresUrl + `/${id}`, {
      lastUpdatedAt: updatedAt,
      feature: {
        displayName,
        technicalName,
        expiresOn,
        description,
        inverted,
        customerIds,
      }
    })
  }

  archiveFeature(featureId: string): Observable<HttpResponse<void>> {
    return this.http.post<HttpResponse<void>>(this.archivedFeaturesUrl, {featureId})
  }
}
