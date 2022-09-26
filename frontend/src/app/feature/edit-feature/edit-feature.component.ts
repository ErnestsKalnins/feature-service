import {Component, OnInit} from '@angular/core';
import {Location} from "@angular/common";
import {ActivatedRoute, Router} from "@angular/router";
import {switchMap} from "rxjs";
import {FeatureService} from "../services/feature.service";
import {Feature} from "../services/feature";

@Component({
  selector: 'app-edit-feature',
  templateUrl: './edit-feature.component.html',
})
export class EditFeatureComponent implements OnInit {
  feature: Feature = {
    id: null,
    technicalName: '',
    displayName: null,
    description: null,
    expiresOn: null,
    inverted: false,
    createdAt: 0,
    updatedAt: 0,
    customerIds: null,
  };

  loading = false;

  constructor(
    private location: Location,
    private route: ActivatedRoute,
    private router: Router,
    private featureService: FeatureService
  ) {
  }

  ngOnInit(): void {
    this.route.paramMap.pipe(
      switchMap(params => {
        return this.featureService.getFeature(params.get('featureId')!);
      })
    ).subscribe(feature => {
      this.feature = feature;
    });
  }

  invert(): void {
    this.feature.inverted = !this.feature.inverted;
  }

  goBack(): void {
    this.location.back();
  }

  getExpiredOnDatetime(): string | null {
    if (this.feature.expiresOn === null) {
      return null
    }
    const utc = new Date(this.feature.expiresOn);
    return new Date(utc.getTime() - utc.getTimezoneOffset() * 60 * 1000)
      .toISOString()
      .slice(0, -1);
  }

  setExpiredOnDatetime(e: any): void {
    this.feature.expiresOn = new Date(e.target.value).valueOf();
  }

  updateFeature(): void {
    this.loading = true;
    const that = this;
    this.featureService.updateFeature(this.feature)
      .subscribe({
        complete() {
          that.router.navigate(['/features']);
        },
        error(e) {
          console.log(e);
          that.loading = false;
        },
      });
  }
}
