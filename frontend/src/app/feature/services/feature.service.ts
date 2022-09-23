import {Injectable} from '@angular/core';
import {HttpClient, HttpResponse} from "@angular/common/http";
import {Observable} from "rxjs";
import {Feature} from "./feature";

@Injectable({
  providedIn: 'root'
})
export class FeatureService {
  private featuresUrl = 'http://localhost:8080/api/v1/features';

  constructor(
    private http: HttpClient,
  ) {
  }

  getFeatures(): Observable<Feature[]> {
    return this.http.get<Feature[]>(this.featuresUrl);
  }

  saveFeature({technicalName, displayName, description, inverted, expiresOn}: Feature): Observable<HttpResponse<void>> {
    const expiresOnRFC3339 = expiresOn === null
      ? null
      : new Date(expiresOn);

    return this.http.post<HttpResponse<void>>(this.featuresUrl, {
      technicalName,
      displayName,
      description,
      inverted,
      expiresOn: expiresOnRFC3339
    });
  }
}